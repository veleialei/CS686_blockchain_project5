package data

import "encoding/json"

type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

// Hello all. For project 3, you have to get an ID for each node before starting or joining a network.
// The registration address is "mc07.cs.usfca.edu:6688/peer". If you "curl" it, you will get an integer as your ID.
// Since our USF servers don't have built-in Golang, the registration server cannot be part of the
// BlockChain network and provide you the current BlockChain during registration. So this server is
// only capable of assigning you a distinct ID.
// Every time you create the first node for your own network, the first node can initiate a new BlockChain.
// The following nodes which join this network will have to (1) Register at TA's server and get an ID.
// (2) Ask the first node for its PeerList. (3) Download the current BlockChain from any peer.

// 1. After a new node is launched, it will go to "mc07.cs.usfca.edu:6688/peer" to register itself, and get an Id(nodeId).
// 2. Then, the node will go to any peer on its PeerList to download the current BlockChain.
// 3. After registration, the node will start to send HeartBeat for every 5~10 seconds.
func NewRegisterData(id int32, peerMapJson string) RegisterData {
	return RegisterData{AssignedId: id, PeerMapJson: peerMapJson}
}

func (data *RegisterData) EncodeToJson() (string, error) {
	ret, err := json.Marshal(data)
	return string(ret), err
}
