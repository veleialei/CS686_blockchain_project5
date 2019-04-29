package data

import (
	"math/rand"

	"../../p1"
)

// PeerMap maps IP Address to its ID. PeerList is a struct containing PeerMap.
// 3 for hops default value
type HeartBeatData struct {
	IfNewBlock    bool   `json:"ifNewBlock"`
	IfUpdateBlock bool   `json:"ifUpdateBlock"`
	CreatorId     string `json:"creatorid"`
	NodeId        int32  `json:"creatorid"`
	BlockJson     string `json:"blockJson"`
	PeerMapJson   string `json:"peerMapJson"`
	Addr          string `json:"addr"`
	Hops          int32  `json:"hops"`
	Secret        string `json:"secret"`
}

var MPT_Q = []string{
	"What is USF in California?\nA) University of San Francisco\nB) University of Florida\n",
	"What is the Snapshot in Spanner?\nA) A set of reads that execute atomically at a single logical point in time across columns, rows, and tables in a database\nB) A photograph taken quickly, typically with a small handheld camera\n",
	"What is FLAG in Tech Company?\nA) Facebook, Linkedin, Amazon, Google\nB) Fendi, Louis Vuitton, Apple, Gucci\n",
	"What can you do if you major in Computer science?\nA) Coding\nB) Fix a broken computer\n",
	"If you have 1 million dollar, what can you buy in bay area?\nA) A small condo\nB) A big single familiar house\n",
}
var MPT_A = []string{
	"A\n",
	"A\n",
	"A\n",
	"A\n",
	"A\n",
}

// Send HeartBeat:
// 1. Every user would hold a PeerList of up to 32 peer nodes. (32 is the number Ethereum uses.)
// The PeerList can temporarily hold more than 32 nodes, but before sending HeartBeats,
// a node will first re-balance the PeerList by choosing the 32 closest peers.
// "Closest peers" is defined by this: Sort all peers' Id, insert SelfId, consider the list as a cycle,
// and choose 16 nodes at each side of SelfId. For example, if SelfId is 10, PeerList is [7, 8, 9, 15, 16],
// then the closest 4 nodes are [8, 9, 15, 16]. HeartBeat is sent to every peer nodes at "/heartbeat/receive".
// 2. For each HeartBeat, a node would randomly decide (this will change in Project 4) if it will create a new block.
// If so, add the block information into HeartBeatData and send the HeartBeatData to others.

// Receive HeartBeat:
// 1. When a node received a HeartBeat, the node will add the sender’s IP address, along with sender’s
// PeerList into its own PeerList. At this time, the number of peers stored in PeerList might exceed 32 and it is ok.
// As described in previously, you don’t have to rebalance every time you receive a HeartBeat.
// Rebalance happens only before you send HeartBeats.
// 2. If the HeartBeatData contains a new block, the node will first check if the previous block exists
//  (the previous block is the block whose hash is the parentHash of the next block).
// 3. If the previous block doesn't exist, the node will ask every peer at "/block/{height}/{hash}" to download that block.
// 4. After making sure previous block exists, insert the block from HeartBeatData to the current BlockChain.
// 5. Since every node only has 32 peers, every peer will forward the new block to all peers according to its PeerList.
// That is to make sure every user in the network would receive the new block. For this project.
// Every HeartBeatData takes 2 hops, which means after a node received a HeartBeatData
// from the original block maker, the remaining hop times is 1.

// 2. NewHeartBeatData() is a normal initial function which creates an instance.
func NewHeartBeatData(ifNewBlock bool, creatorId string, nodeId int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	data := HeartBeatData{IfNewBlock: ifNewBlock, CreatorId: creatorId, NodeId: nodeId, BlockJson: blockJson, PeerMapJson: peerMapJson, Addr: addr, Hops: 2}
	return data
}

// PrepareHeartBeatData() is used when you want to send a HeartBeat to other peers.
// PrepareHeartBeatData would first create a new instance of HeartBeatData,
// then decide whether or not you will create a new block and send the new block to other peers.
func PrepareHeartBeatData(sbc *SyncBlockChain, id string, nodeId int32, peerMapBase64 string, addr string) HeartBeatData {
	data := NewHeartBeatData(false, id, nodeId, "", peerMapBase64, addr)
	return data
}

//generate a random MPT
func GenMPT(content string, react string) p1.MerklePatriciaTrie {
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	i := rand.Intn(5)
	if content == "" {
		mpt.Insert("content", MPT_Q[i])
		mpt.Insert("react", MPT_A[i])
	} else {
		mpt.Insert("content", content)
		mpt.Insert("react", react)
	}

	return mpt
}
