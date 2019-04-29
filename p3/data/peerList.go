package data

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// PeerMap maps IP Address to its ID. PeerList is a struct containing PeerMap.
type PeerList struct {
	selfId    int32
	peerMap   map[string]int32
	maxLength int32
	mux       sync.Mutex
}

type Peer struct {
	Addr string `json:"addr"`
	Id   int32  `json:"id"`
}

func NewPeerList(id int32, maxLength int32) PeerList {
	return PeerList{selfId: id, peerMap: make(map[string]int32), maxLength: maxLength, mux: sync.Mutex{}}
}

func (peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	if id != peers.selfId {
		peers.peerMap[addr] = id
	}
	peers.mux.Unlock()
}

func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.peerMap, addr)
	peers.mux.Unlock()
}

// https://golang.org/pkg/container/ring/
// for this function, if the size of peers is less than 32,
// there will be no rebalancing needed. Is that correct?
func (peers *PeerList) Rebalance() {
	peers.mux.Lock()
	fmt.Println("Locked")
	size := len(peers.peerMap)
	if size <= int(peers.maxLength) {
		peers.mux.Unlock()
		fmt.Println("UnLocked")
		return
	}

	type kv struct {
		Key   string
		Value int32
	}

	var ss []kv
	for k, v := range peers.peerMap {
		ss = append(ss, kv{k, v})
	}
	ss = append(ss, kv{"self", peers.selfId})

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value < ss[j].Value
	})

	pos := 0

	for ss[pos].Value != peers.selfId {
		pos++
	}

	fmt.Println(pos)
	maxLen := int(peers.maxLength)
	// [1](end) 2 3 [4 5 ] 6 [7]   pos = 6, len(ss) = 7 maxlen/2 = 2
	end := pos + maxLen/2 + 1 // 7

	if end >= len(ss) {
		end = end - len(ss) //  0
	}

	start := pos - maxLen/2 - 1 // 1 - 2 = -1, ed = 1 + 2 = 3
	// [1] 2 [3 4] 5  6 [7]   pos = 1  len(ss) = 7 maxlen/2 = 2
	if start < 0 {
		start = len(ss) + start // 3 - 6
	}

	if start > end {
		tmp := end
		end = start
		start = tmp
	}

	for i := start; i <= end; i++ {
		delete(peers.peerMap, ss[i].Key)
		fmt.Println("Delete: ", ss[i].Value)
	}

	peers.mux.Unlock()
	fmt.Println("UnLocked")
}

//???
// 1. Show() shows all addresses and their corresponding IDs.
// For example, it returns "This is PeerMap: \n addr=127.0.0.1, id=1".
func (peers *PeerList) Show() string {
	ret := "This is PeerMap: \n"
	for k, v := range peers.peerMap {
		ret = ret + "addr=" + k + ", id=" + string(v) + "\n"
	}
	fmt.Printf(ret)
	return ret
}

func (peers *PeerList) Register(id int32) {
	peers.selfId = id
	fmt.Printf("SelfId=%v\n", id)
}

func (peers *PeerList) Copy() map[string]int32 {
	copy := map[string]int32{}
	for k, v := range peers.peerMap {
		copy[k] = v
	}
	return copy
}

func (peers *PeerList) GetSelfId() int32 {
	return peers.selfId
}

func (peers *PeerList) PeerMapToJson() (string, error) {
	ret, err := json.Marshal(peers)
	return string(ret), err
}

// 3. The "PeerMapJson" in HeartBeatData is the JSON format of "PeerList.peerMap".
// It is the result of "PeerList.PeerMapToJSON()" function. Sorry for the confused
// argument name "PeerMapBase64" in PerpareHeartBeatData().

// 4. There might not be a pre-test for project 3. You can launch the whole system with several nodes,
// and check if all functionalities work fine such as heartBeats, initializations, downloading and uploading.

// 2. InjectPeerMapJson() inserts every entries(every <addr, id> pair) of the parameter "peerMapJsonStr"
// into your own PeerMap, except the entry whose addres is your own local address.
func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	var newMap []Peer
	err := json.Unmarshal([]byte(peerMapJsonStr), &newMap)
	if err != nil {
		return
	}
	for _, v := range newMap {
		if v.Id != peers.selfId {
			peers.Add(v.Addr, v.Id)
		}
	}
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(peers)
	fmt.Println(expected)
	fmt.Println(reflect.DeepEqual(peers, expected))

	fmt.Println("==================================")

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(peers)
	fmt.Println(expected)
	fmt.Println(reflect.DeepEqual(peers, expected))

	fmt.Println("==================================")

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(peers)
	fmt.Println(expected)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
