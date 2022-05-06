package backend

import (
	"container/heap"
	"xfsgo/p2p/discover"
)

// peers grade system
const (
	primaryPeer = iota
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

//UpdatereFusePeer In P2P network, deleting local peer connection has no effect, so we need to set its level to reject
func (p *peerLevel) UpdatereFusePeer() {
	p.Level = tertiaryPeer
}

type peerslist []*peerLevel

func (ps *peerslist) SortLevelAndHeightbasePeer() peerslist {
	primaryPeerList := ps.ExcellentPeersQueue()

	var result peerslist
	sortpsHeight := peersSortHeightList(*primaryPeerList)

	heap.Init(&sortpsHeight)

	for sortpsHeight.Len() > 0 {
		result = append(result, heap.Pop(&sortpsHeight).(*peerLevel))
	}
	return result
}

func (ps *peerslist) Remove(id discover.NodeId) {
	newps := make(peerslist, len(*ps)-1)
	for _, v := range *ps {
		if v.NodeId.String() != id.String() {
			newps = append(newps, v)
		}
	}
	ps = &newps
}

// AllPeers Get excellent and passed peers
// Exclude reject level peer
func (ps *peerslist) AllPeers() *peerslist {
	result := make(peerslist, 0)
	result = append(result, *ps.FusePeersQueue()...)
	result = append(result, *ps.ExcellentPeersQueue()...)
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

// type peerSortLevelList []*peerLevel

// func (ps peerSortLevelList) Len() int {
// 	return len(ps)
// }

// func (ps peerSortLevelList) Less(i, j int) bool {
// 	return ps[i].Level < ps[j].Level
// }

// func (ps peerSortLevelList) Swap(i, j int) {
// 	ps[i], ps[j] = ps[j], ps[i]
// }

// func (ps *peerSortLevelList) Push(x interface{}) {
// 	*ps = append(*ps, x.(*peerLevel))
// }

// func (ps *peerSortLevelList) Pop() interface{} {
// 	old := *ps
// 	n := len(old)
// 	x := old[n-1]
// 	*ps = old[0 : n-1]
// 	return x
// }

type peersSortHeightList []*peerLevel

func (ps peersSortHeightList) Len() int {
	return len(ps)
}

func (ps peersSortHeightList) Less(i, j int) bool {
	return ps[i].Syncpeer.Height() > ps[j].Syncpeer.Height()
}

func (ps peersSortHeightList) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func (ps *peersSortHeightList) Push(x interface{}) {
	*ps = append(*ps, x.(*peerLevel))
}

func (ps *peersSortHeightList) Pop() interface{} {
	old := *ps
	n := len(old)
	x := old[n-1]
	*ps = old[0 : n-1]
	return x
}
