package discover

import (
	"bytes"
	"container/list"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
	"xfsgo/common/ahash"
	"xfsgo/common/rawencode"
	"xfsgo/crypto"
	"xfsgo/log"
	"xfsgo/p2p/nat"

	"github.com/sirupsen/logrus"
)

const Version = 1

// Errors
var (
	errExpired          = errors.New("expired")
	errBadVersion       = errors.New("version mismatch")
	errUnsolicitedReply = errors.New("unsolicited reply")
	errUnknownNode      = errors.New("unknown node")
	errTimeout          = errors.New("RPC timeout")
	errClockWarp        = errors.New("reply deadline too far in the future")
	errClosed           = errors.New("socket closed")
)

// Timeouts
const (
	respTimeout = 500 * time.Millisecond
	expiration  = 20 * time.Second

	refreshInterval = 1 * time.Hour
)

// RPC Packet Type
const (
	PING  = 0x00 // Request
	PONG  = 0x01 // Response
	FIND  = 0x02 // Request (Only UDP)
	NODES = 0x03 // Response (Only UDP)
	MSG   = 0x04 // Message
)

func makeEndpoint(addr *net.UDPAddr, tcpPort uint16) rpcEndpoint {
	ip := addr.IP.To4()
	if ip == nil {
		ip = addr.IP.To16()
	}
	return rpcEndpoint{IP: ip, UDP: uint16(addr.Port), TCP: tcpPort}
}

func nodeFromRPC(rn rpcNode) (n *Node, valid bool) {
	// TODO: don't accept localhost, LAN addresses from internet hosts
	// TODO: check public key is on secp256k1 curve
	if rn.IP.IsMulticast() || rn.IP.IsUnspecified() || rn.UDP == 0 {
		return nil, false
	}
	return newNode(rn.IP, rn.TCP, rn.UDP, rn.ID), true
}

func nodeToRPC(n *Node) rpcNode {
	return rpcNode{ID: n.ID, IP: n.IP, UDP: n.UDP, TCP: n.TCP}
}

type packet interface {
	handle(t *udp, from *net.UDPAddr, fromID NodeId, targetID NodeId) error
}

type conn interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error)
	Close() error
	LocalAddr() net.Addr
}

// udp implements the RPC protocol.
type udp struct {
	//logger log.Logger
	conn        conn
	priv        *ecdsa.PrivateKey
	ourEndpoint rpcEndpoint

	addpending chan *pending
	gotreply   chan reply

	closing chan struct{}
	log     log.Logger
	*Table
}

// pending represents a pending reply.
//
// some implementations of the protocol wish to send more than one
// reply packet to findnode. in general, any neighbors packet cannot
// be matched up with a specific findnode packet.
//
// our implementation handles this by storing a callback function for
// each pending reply. incoming packets from a node are dispatched
// to all the callback functions for that node.
type pending struct {
	// these fields must match in the reply.
	from  NodeId
	ptype byte

	// time when the request must complete
	deadline time.Time

	// callback is called when a matching reply arrives. if it returns
	// true, the callback is removed from the pending reply queue.
	// if it returns false, the reply is considered incomplete and
	// the callback will be invoked again for the next matching reply.
	callback func(resp interface{}) (done bool)

	// errc receives nil when the callback indicates completion or an
	// error if no further reply is received within the timeout.
	errc chan<- error
}

type reply struct {
	from  NodeId
	ptype byte
	data  interface{}
	// loop indicates whether there was
	// a matching request by sending on this channel.
	matched chan<- bool
}

// ListenUDP returns a new table that listens for UDP packets on laddr.
func ListenUDP(priv *ecdsa.PrivateKey, laddr string, nodeDBPath string, mapper nat.Mapper, log log.Logger) (*Table, error) {
	addr, err := net.ResolveUDPAddr("udp", laddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	tab, _ := newUDP(priv, conn, nodeDBPath, mapper, log)
	return tab, nil
}
func NewUDP(priv *ecdsa.PrivateKey, c conn, nodeDBPath string, mapper nat.Mapper, log log.Logger) (*Table, *udp) {
	return newUDP(priv, c, nodeDBPath, mapper, log)
}
func newUDP(priv *ecdsa.PrivateKey, c conn, nodeDBPath string, mapper nat.Mapper, log log.Logger) (*Table, *udp) {
	udp := &udp{
		log:        log,
		conn:       c,
		priv:       priv,
		closing:    make(chan struct{}),
		gotreply:   make(chan reply),
		addpending: make(chan *pending),
	}
	realaddr := c.LocalAddr().(*net.UDPAddr)
	if mapper != nil && !realaddr.IP.IsLoopback() {
		go nat.Map(mapper, udp.closing, "udp", realaddr.Port, realaddr.Port, "xlibp2p discovery")
	} else if mapper != nil {
		if ext, err := mapper.ExternalIP(); err == nil {
			realaddr = &net.UDPAddr{IP: ext, Port: realaddr.Port}
		}
	}
	udp.ourEndpoint = makeEndpoint(realaddr, uint16(realaddr.Port))
	udp.Table = newTable(udp, PubKey2NodeId(priv.PublicKey), realaddr, nodeDBPath, log)
	go udp.loop()
	go udp.readLoop()
	return udp.Table, udp
}

func (t *udp) close() {
	close(t.closing)
	_ = t.conn.Close()
	// TODO: wait for the loops to end.
}

// ping sends a ping message to the given node and waits for a reply.
func (t *udp) ping(fromId, toid NodeId, toaddr *net.UDPAddr) error {
	// TODO: maybe check for ReplyTo field in callback to measure RTT
	errc := t.pending(toid, PONG, func(interface{}) bool { return true })
	_ = t.sendN(fromId, toaddr, PING, ping{
		Version:    Version,
		From:       t.ourEndpoint,
		To:         makeEndpoint(toaddr, 0), // TODO: maybe use known TCP port from DB
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}, toid)
	return <-errc
}

func (t *udp) waitping(from NodeId) error {
	return <-t.pending(from, PING, func(interface{}) bool { return true })
}

// findnode sends a findnode request to the given node and waits until
// the node has sent up to k neighbors.
func (t *udp) findnode(fromId NodeId, toid NodeId, toaddr *net.UDPAddr, target NodeId) ([]*Node, error) {
	nodes := make([]*Node, 0, bucketSize)
	nreceived := 0
	errc := t.pending(toid, NODES, func(r interface{}) bool {
		reply := r.(*neighbors)
		for _, rn := range reply.Nodes {
			nreceived++
			if n, valid := nodeFromRPC(rn); valid {
				nodes = append(nodes, n)
			}
		}
		return nreceived >= bucketSize
	})
	_ = t.sendN(fromId, toaddr, FIND, findnode{
		Target:     target,
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}, toid)
	err := <-errc
	return nodes, err
}

// pending adds a reply callback to the pending reply queue.
// see the documentation of type pending for a detailed explanation.
func (t *udp) pending(id NodeId, ptype byte, callback func(interface{}) bool) <-chan error {
	ch := make(chan error, 1)
	p := &pending{from: id, ptype: ptype, callback: callback, errc: ch}
	select {
	case t.addpending <- p:
		// loop will handle it
	case <-t.closing:
		ch <- errClosed
	}
	return ch
}

func (t *udp) handleReply(targetID NodeId, ptype byte, req packet) bool {
	matched := make(chan bool)
	select {
	case t.gotreply <- reply{targetID, ptype, req, matched}:
		// loop will handle it
		return <-matched
	case <-t.closing:
		return false
	}
}

// loop runs in its own goroutin. it keeps track of
// the refresh timer and the pending reply queue.
func (t *udp) loop() {
	var (
		plist       = list.New()
		timeout     = time.NewTimer(0)
		nextTimeout *pending // head of plist when timeout was last reset
		refresh     = time.NewTicker(refreshInterval)
	)
	<-timeout.C // ignore first timeout
	defer refresh.Stop()
	defer timeout.Stop()

	resetTimeout := func() {
		if plist.Front() == nil || nextTimeout == plist.Front().Value {
			return
		}
		// Start the timer so it fires when the next pending reply has expired.
		now := time.Now()
		for el := plist.Front(); el != nil; el = el.Next() {
			nextTimeout = el.Value.(*pending)
			//t.logger.Debugf("udp loop resetTimeout from: %s, packet type: %x, now: %s, next timeout: %s", nextTimeout.from, nextTimeout.ptype, now.Format(time.RFC3339), nextTimeout.deadline.Format(time.RFC3339))
			if dist := nextTimeout.deadline.Sub(now); dist < 2*respTimeout {
				//t.logger.Debugf("udp loop from: %s, packet type: %x, resetTimeout: %s",nextTimeout.from,nextTimeout.ptype, dist)
				timeout.Reset(dist)
				return
			}
			// Remove pending replies whose deadline is too far in the
			// future. These can occur if the system clock jumped
			// backwards after the deadline was assigned.
			nextTimeout.errc <- errClockWarp
			plist.Remove(el)
		}
		nextTimeout = nil
		timeout.Stop()
	}

	for {
		resetTimeout()

		select {
		case <-refresh.C:
			go t.refresh()

		case <-t.closing:
			for el := plist.Front(); el != nil; el = el.Next() {
				el.Value.(*pending).errc <- errClosed
			}
			return

		case p := <-t.addpending:
			p.deadline = time.Now().Add(respTimeout)

			//t.logger.Debugf("udp loop receive addpending chan from: %s, type: %x, deadline: %s",p.from,p.ptype, p.deadline.Format(time.RFC3339))
			plist.PushBack(p)

		case r := <-t.gotreply:
			//t.logger.Debugf("udp loop receive gotreply chan from: %s, type: %x", r.from, r.ptype)
			var matched bool
			for el := plist.Front(); el != nil; el = el.Next() {
				p := el.Value.(*pending)
				if p.from == r.from && p.ptype == r.ptype {
					matched = true
					// Remove the matcher if its callback indicates
					// that all replies have been received. This is
					// required for packet types that expect multiple
					// reply packets.
					if p.callback(r.data) {
						p.errc <- nil
						plist.Remove(el)
					}
				}
			}
			r.matched <- matched

		case now := <-timeout.C:
			nextTimeout = nil
			// Notify and remove callbacks whose deadline is in the past.
			for el := plist.Front(); el != nil; el = el.Next() {
				p := el.Value.(*pending)
				//t.logger.Debugf("udp loop timeout tick, now: %s, p.deadline: %s", now, p.deadline)
				if now.After(p.deadline) || now.Equal(p.deadline) {
					p.errc <- errTimeout
					plist.Remove(el)
				}
			}
		}
	}
}
func (t *udp) sendN(fromId NodeId, toaddr *net.UDPAddr, ptype byte, req interface{}, to NodeId) error {
	packet, err := encodePacketN(fromId, t.priv, ptype, req, to)
	if err != nil {
		return err
	}
	if _, err = t.conn.WriteToUDP(packet, toaddr); err != nil {
		logrus.Infof("UDP send failed:%v", err)
	}
	return err
}

func encodePacketN(fromId NodeId, privateKey *ecdsa.PrivateKey, ptype byte, req interface{}, remote NodeId) ([]byte, error) {
	bytePacket := new(bytes.Buffer)
	bytePacket.WriteByte(Version)
	bytePacket.WriteByte(ptype)
	bytePacket.Write(fromId[:])
	bytePacket.Write(remote[:])

	bs, err := rawencode.Encode(req)
	if err != nil {
		return nil, err
	}
	datahash := ahash.SHA256(bs)
	signed, err := crypto.ECDSASign(datahash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("can't sign discv4 packet%v", err)
	}
	bytePacket.Write(signed)

	var dataLen uint32 = uint32(len(bs))
	var byteLen []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(byteLen, dataLen)
	bytePacket.Write(byteLen)
	bytePacket.Write(bs)
	return bytePacket.Bytes(), nil
}

// readLoop runs in its own goroutine. it handles incoming UDP packets.
func (t *udp) readLoop() {
	defer func() {
		if err := t.conn.Close(); err != nil {
			logrus.Debugf("UDP Close:%v", err)
		}
	}()
	// Discovery packets are defined to be no larger than 1280 bytes.
	// Packets larger than this size will be cut at the end and treated
	// as invalid because their hash won't match.
	buf := make([]byte, 1280)
	for {
		nbytes, from, err := t.conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		err = t.handlePacketN(from, buf[:nbytes])
		if err != nil {
			continue
		}
	}
}
func (t *udp) handlePacketN(from *net.UDPAddr, buf []byte) error {
	buffer := bytes.NewBuffer(buf)
	packet, fromID, targetID, err := decodePacketN(buffer, t.self.ID)
	if err != nil {
		logrus.Debugf("bad packet:%v", err)
		return err
	}

	if err = packet.handle(t, from, fromID, targetID); err != nil {
		logrus.Debugf("packet.handle:%v", err)
	}

	return err
}

type packetHeadN struct {
	version  uint8    // Version
	packType uint8    // Type
	from     [32]byte // From
	to       [32]byte // To
	sign     [65]byte // Signature
	len      uint32   // Payload Length
}

const headerLen = 135

func decodePacketHeadN(reader io.Reader) (*packetHeadN, int, error) {
	// 0000
	var hbytes [headerLen]byte
	last := 0
	for last < headerLen {
		var buf [headerLen]byte
		n, err := reader.Read(buf[:])
		if err != nil && err == io.EOF {
			return nil, last, err
		}
		copy(hbytes[last:last+n], buf[:n])
		last += n
	}
	ph := new(packetHeadN)
	version, mType, from, to, sign, dataLen := hbytes[0], hbytes[1], hbytes[2:34], hbytes[34:66], hbytes[66:131], hbytes[131:135]
	ph.version = version
	ph.packType = mType
	copy(ph.sign[:], sign[:])
	copy(ph.from[:], from[:])
	copy(ph.to[:], to[:])

	ph.len = binary.BigEndian.Uint32(dataLen)
	return ph, last, nil
}

func recoverNodeId(hash, sig []byte) (NodeId, error) {
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return NodeId{}, err
	}
	return PubKey2NodeId(*pubKey), nil
}
func decodePacketN(reader io.Reader, self NodeId) (packet, NodeId, NodeId, error) {
	h, _, err := decodePacketHeadN(reader)
	if err != nil {
		return nil, NodeId{}, NodeId{}, err
	}
	if !bytes.Equal(self[:], h.to[:]) {
		return nil, NodeId{}, NodeId{}, fmt.Errorf("id not match")
	}

	var data = make([]byte, h.len)
	last := uint32(0)
	for last < h.len {
		var buf [^uint8(0)]byte
		n, err := reader.Read(buf[:])
		if err != nil && err == io.EOF {
			return nil, NodeId{}, NodeId{}, err
		}
		copy(data[last:int(last)+n], buf[:n])
		last += uint32(n)
	}
	datahash := ahash.SHA256(data)

	var from [32]byte // FromID
	copy(from[:], h.from[:])
	nid, err := recoverNodeId(datahash, h.sign[:])
	if err != nil {
		return nil, NodeId{}, nid, err
	}
	var req packet
	switch ptype := h.packType; ptype {
	case PING:
		req = new(ping)
	case PONG:
		req = new(pong)
	case FIND:
		req = new(findnode)
	case NODES:
		req = new(neighbors)
	default:
		return nil, from, nid, fmt.Errorf("unknown type: %d", ptype)
	}
	err = rawencode.Decode(data, req)
	return req, from, nid, err
}
