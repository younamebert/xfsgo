package backend

import (
	"container/heap"
	"xfsgo/p2p/discover"
)

// peers grade system
const (
	primaryPeer = iota
	secondaryPeer
	tertiaryPeer
)

type peerLevel struct {
	Level    int
	NodeId   discover.NodeId
	Syncpeer syncpeer
}

func newpeerLevel(id discover.NodeId, p syncpeer) *peerLevel {
	return &peerLevel{
		NodeId:   id,
		Level:    primaryPeer,
		Syncpeer: p,
	}
}

func (p *peerLevel) UpdateExcellentPeer() {
	p.Level = primaryPeer
}

func (p *peerLevel) UpdateOrdinaryPeer() {
	p.Level = secondaryPeer
}

//UpdatereFusePeer In P2P network, deleting local peer connection has no effect, so we need to set its level to reject
func (p *peerLevel) UpdatereFusePeer() {
	p.Level = tertiaryPeer
}

type peerslist []*peerLevel

func (ps peerslist) Len() int {
	return len(ps)
}

func (ps peerslist) Less(i, j int) bool {
	return ps[i].Level < ps[j].Level
}

func (ps peerslist) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func (ps *peerslist) Push(x interface{}) {
	*ps = append(*ps, x.(*peerLevel))
}

func (ps *peerslist) Pop() interface{} {
	old := *ps
	n := len(old)
	x := old[n-1]
	*ps = old[0 : n-1]
	return x
}

func (ps *peerslist) BasePeers() {
	heap.Init(ps)
	for ps.Len() > 0 {
		heap.Pop(ps)
	}
}

// AllPeers Get excellent and passed peers
// Exclude reject level peer
func (ps *peerslist) AllPeers() *peerslist {
	result := make(peerslist, 0)
	for _, v := range *ps {
		if v.Level < tertiaryPeer {
			result = append(result, v)
		}
	}
	return &result
}

// FusePeersQueue Get reject level peer queue
func (ps *peerslist) FusePeersQueue() *peerslist {
	result := make(peerslist, 0)
	for _, v := range *ps {
		if v.Level == tertiaryPeer {
			result = append(result, v)
		}
	}
	return &result
}

// OrdinaryPeersQueue Get normal peer level queue
func (ps *peerslist) OrdinaryPeersQueue() *peerslist {
	result := make(peerslist, 0)
	for _, v := range *ps {
		if v.Level == secondaryPeer {
			result = append(result, v)
		}
	}
	return &result
}

// ExcellentPeersQueue Get excellent peer level queue
func (ps *peerslist) ExcellentPeersQueue() *peerslist {
	result := make(peerslist, 0)
	for _, v := range *ps {
		if v.Level == primaryPeer {
			result = append(result, v)
		}
	}
	return &result
}
