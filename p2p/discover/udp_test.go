package discover

import (
	"bytes"
	"net"
	"testing"
	"xfsgo/common/ahash"
	"xfsgo/common/rawencode"
	"xfsgo/crypto"
)

func Test_udp_encodePacketN(t *testing.T) {
	dataType := uint8(1)
	data := "hello"
	dataraw, err := rawencode.Encode(data)
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.MustGenPrvKey()
	datahash := ahash.SHA256(dataraw)
	wantSign, err := crypto.ECDSASign(datahash, key)
	if err != nil {
		t.Fatal(err)
	}

	fromKey := crypto.MustGenPrvKey()
	fromId := PubKey2NodeId(fromKey.PublicKey)

	remoteKey := crypto.MustGenPrvKey()
	wantRemote := PubKey2NodeId(remoteKey.PublicKey)
	encodedraw, err := encodePacketN(fromId, key, dataType, data, wantRemote)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("raw: %x", encodedraw)
	if encodedraw[0] != dataType {
		t.Fatal("invalid packet type")
	}
	if !bytes.Equal(wantSign, encodedraw[1:66]) {
		t.Fatal("invalid packet sign")
	}
	if !bytes.Equal(wantRemote[:], encodedraw[66:130]) {
		t.Fatal("invalid packet remote")
	}
	if len(dataraw) != int(encodedraw[130]) {
		t.Fatal("invalid packet data length")
	}
	if !bytes.Equal(dataraw, encodedraw[131:]) {
		t.Fatal("invalid packet data")
	}
}
func Test_udp_decodePacketHeadN(t *testing.T) {
	dataType := uint8(1)
	data := "hello"
	dataraw, err := rawencode.Encode(data)
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.MustGenPrvKey()
	datahash := ahash.SHA256(dataraw)
	wantSign, err := crypto.ECDSASign(datahash, key)
	if err != nil {
		t.Fatal(err)
	}

	fromKey := crypto.MustGenPrvKey()
	fromId := PubKey2NodeId(fromKey.PublicKey)
	remoteKey := crypto.MustGenPrvKey()
	wantRemote := PubKey2NodeId(remoteKey.PublicKey)
	encodedraw, err := encodePacketN(fromId, key, dataType, data, wantRemote)
	if err != nil {
		t.Fatal(err)
	}
	buffer := bytes.NewBuffer(encodedraw)
	ph, _, err := decodePacketHeadN(buffer)
	if err != nil {
		t.Fatal(err)
	}
	if dataType != ph.packType {
		t.Fatal("invalid packet type")
	}
	if !bytes.Equal(wantSign, ph.sign[:]) {
		t.Fatal("invalid packet sign")
	}
	if len(dataraw) != int(ph.len) {
		t.Fatal("invalid packet data length")
	}
}
func Test_udp_decodePacketN(t *testing.T) {
	key := crypto.MustGenPrvKey()
	wantNodeId := PubKey2NodeId(key.PublicKey)
	fromAddr := &net.UDPAddr{
		IP:   net.IPv4(0xff, 0xff, 0xff, 0xff),
		Port: 0xffff,
	}
	toAddr := &net.UDPAddr{
		IP:   net.IPv4(0xff, 0xff, 0xff, 0xff),
		Port: 0xffff,
	}
	fromEndpoint := makeEndpoint(fromAddr, uint16(fromAddr.Port))
	toEndpoint := makeEndpoint(toAddr, uint16(toAddr.Port))
	pack := ping{
		Version:    1,
		From:       fromEndpoint,
		To:         toEndpoint,
		Expiration: 100,
	}
	fromKey := crypto.MustGenPrvKey()
	fromId := PubKey2NodeId(fromKey.PublicKey)
	remoteKey := crypto.MustGenPrvKey()
	wantRemote := PubKey2NodeId(remoteKey.PublicKey)
	encodedraw, err := encodePacketN(fromId, key, PING, pack, wantRemote)
	if err != nil {
		t.Fatal(err)
	}
	buffer := bytes.NewBuffer(encodedraw)
	p, _, n, err := decodePacketN(buffer, wantRemote)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("invalid parse packet")
	}
	if !bytes.Equal(wantNodeId[:], n[:]) {
		t.Fatal("invalid packet node id")
	}
}
