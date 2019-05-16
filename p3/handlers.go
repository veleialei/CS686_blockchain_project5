package p3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../p2"
	"./data"
	"github.com/gorilla/mux"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var SELF_ADDR = "http://localhost:6680"

var ID int32 = 123
var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool
var Hex = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}

// ssh -L 6688:mc07.cs.usfca.edu:6688 <username>@stargate.cs.usfca.ed

// Init():
// Create SyncBlockChain and PeerList instances.
func Init() {
	// This function will be executed before everything else.
	// Do some initialization here.
	fmt.Println("Initing")
	SBC = data.NewBlockChain()
	Peers = data.NewPeerList(0, 32)
	if ID == 123 {
		mpt := data.GenMPT("I want to start", "OK")
		rank := make(map[string]int32)
		rank["123"] = 1
		SBC.GenBlock(mpt, rank, "123")
	}
}

// Start():
// Get an ID from TA's server, download the BlockChain from your own first node,
// use "go StartHeartBeat()" to start HeartBeat loop.
func Start(w http.ResponseWriter, r *http.Request) {
	// 1. After a new node is launched, it will go to
	// "mc07.cs.usfca.edu:6688/peer" to register itself, and get an Id(nodeId).
	if !ifStarted {
		fmt.Println("Starting")

		Init()

		Register()

		Download()

		go StartHeartBeat()

		ifStarted = true
	}
}

// Show():
// Shows the PeerMap and the BlockChain.
func Show(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register():
// Go to TA's server, get an ID.
func Register() {
	fmt.Println("Register")
	Peers.Register(ID)
	data.NewRegisterData(ID, "")
}

// RegisterTA():
// Go to TA's server, get an ID.
func RegisterTA() {
	response, err := http.Get(REGISTER_SERVER)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	id, err := strconv.Atoi(string(body))

	if err != nil {
		log.Fatal(err)
	}
	Peers.Register(int32(id))

	data.NewRegisterData(int32(id), "")
}

// Download():
// Download the current BlockChain from your own first node(can be hardcoded).
// It's ok to use this function only after launching a new node. You may not need it after node starts heartBeats.
func Download() {
	fmt.Println("Download")
	var peer data.Peer
	peer.Id = Peers.GetSelfId()
	peer.Addr = SELF_ADDR
	jsonObj, err := json.Marshal(peer)

	req, err := http.NewRequest("POST", BC_DOWNLOAD_SERVER, bytes.NewBuffer(jsonObj))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("err in GET TA server")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
		return
	}

	fmt.Println("GET BODY: " + string(body))
	SBC.UpdateEntireBlockChain(string(body))
}

// Upload():
// Return the BlockChain's JSON. And add the remote peer into the PeerMap.
func Upload(w http.ResponseWriter, r *http.Request) {
	//add the remote's id and addr
	fmt.Println("Upload")
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var peer data.Peer
	err = json.Unmarshal([]byte(body), &peer)
	if err != nil {
		data.PrintError(err, "Upload")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("parse request body fail"))
		return
	}
	Peers.Add(peer.Addr, peer.Id)

	//return the blockchain's json
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		data.PrintError(err, "Upload")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error when get block chain"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(blockChainJson))
}

// /block/{height}/{hash}
// Method: GET
// Response: If you have the block, return the JSON string of the specific block;
// if you don't have the block, return HTTP 204: StatusNoContent;
// if there's an error, return HTTP 500: InternalServerError.
// Description: Return JSON string of a specific block to the downloader.
// Ask another server to return a block of certain height and hash
// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Upload Block")
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}
	vars := mux.Vars(r)
	height, err := strconv.Atoi(vars["height"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: InternalServerError"))
		return
	}

	hash := vars["hash"]
	block, valid := SBC.GetBlock(int32(height), hash)

	if !valid {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("HTTP 204: StatusNoContent"))
		return
	}

	// fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(block.EncodeToJson()))
}

// HeartBeatReceive(): Alter this function so that when it receives a HeartBeatData with a new block,
// it verifies the nonce as described above. TODO
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}

	fmt.Println("HeartBeatReceive")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot read body"))
		return
	}
	// 1. When a node received a HeartBeat, the node will add the sender’s IP address,
	// along with sender’s PeerList into its own PeerList. No rebalance here
	var heartBeatData data.HeartBeatData
	err = json.Unmarshal([]byte(body), &heartBeatData)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("body is not a valid json format of heartbeat data"))
		return
	}

	Peers.Add(heartBeatData.Addr, ID)
	Peers.InjectPeerMapJson(heartBeatData.PeerMapJson, heartBeatData.Addr)
	// 2. If the HeartBeatData contains a new block, the node will first check
	// if the previous block exists (the previous block is the block whose hash
	// is the parentHash of the next block).
	//TODO!!!!!!!!!!!!!!
	if heartBeatData.IfNewBlock || heartBeatData.IfUpdateBlock {
		block := p2.DecodeFromJson(heartBeatData.BlockJson)
		found := SBC.CheckParentHash(*block)
		if !found {
			// 3. If the previous block doesn't exist, the node will ask every peer
			// at "/block/{height}/{hash}" to download that block.
			// 4. After making sure previous block exists, insert the block from HeartBeatData to the current BlockChain.
			parentStr := AskForBlock(block.Header.Height-1, block.Header.ParentHash)
			if len(parentStr) == 0 {
				fmt.Println("FORWARD/ cannot find one of the parent block")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("cannot find the parent block"))
				return
			} else {
				found = true
			}
		} else {
			if heartBeatData.IfNewBlock {
				parentBlock := SBC.GetParentBlock(*block)
				if parentBlock.VerifySecret(heartBeatData.CreatorId, heartBeatData.Secret) {
					SBC.Insert(*block)
					fmt.Println("FORWARD/ new block inserted: ", block)
				} else {
					fmt.Println("verification failed")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("verification failed cannot insert"))
					return
				}
			} else {
				success := SBC.UpdateBlock(*block, heartBeatData.CreatorId)
				if !success {
					fmt.Println("verification failed")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("verification failed cannot update"))
					return
				}
			}
		}
	}
	// 5. Since every node only has 32 peers, every peer will forward the new block to
	// all peers according to its PeerList. That is to make sure every user
	// in the network would receive the new block. For this project.
	// Every HeartBeatData takes 2 hops, which means after a node received a
	// HeartBeatData from the original block maker, the remaining hop times is 1.
	heartBeatData.Hops--
	if heartBeatData.Hops > 0 {
		ForwardHeartBeat(heartBeatData)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

// AskForBlock will be called in HeartBeatReceive,
// in AskForBlock you will call http get to
// /localhost:port/block/{height}/{hash} (UploadBlock) to get the Block
// Loop through all peers in local PeerMap to download a block. As soon as one peer returns the block, stop the loop.

// AskForBlock(): Update this function to recursively ask for all the missing predesessor blocks instead of only the parent block.
func AskForBlock(height int32, hash string) string {
	peerMap := Peers.Copy()
	for k := range peerMap {
		body, code := AskForBlockHttpRequest(k, height, hash)
		if code == 200 {
			parentBlock := p2.DecodeFromJson(body)
			parentHash := parentBlock.Header.ParentHash
			parentHeight := parentBlock.Header.Height
			if parentBlock.Header.ParentHash == "genesis" {
				SBC.Insert(*parentBlock)
				return "success"
			}
			block := SBC.GetParentBlock(*parentBlock)
			if block.Header.Height > 1 {
				result := AskForBlock(parentHeight-1, parentHash)
				if result == "success" {
					SBC.Insert(*parentBlock)
				}
				return result
			} else {
				return "success"
			}

		}
	}
	return ""
}

func AskForBlockHttpRequest(addr string, height int32, hash string) (string, int) {
	response, err := http.Get(addr + "/block/" + string(height) + "/" + hash)
	if err != nil {
		fmt.Println(err)
	}
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			return string(body), -1
		}
		return string(body), 200
	}
	return "", -1
}

// Send HeartBeat:
// 1. Every user would hold a PeerList of up to 32 peer nodes.
// (32 is the number Ethereum uses.) The PeerList can temporarily hold more than 32 nodes,
// but before sending HeartBeats, a node will first re-balance the PeerList

// by choosing the 32 closest peers. "Closest peers" is defined by this: Sort all peers' Id,
// insert SelfId, consider the list as a cycle, and choose 16 nodes at each side of SelfId.
// For example, if SelfId is 10, PeerList is [7, 8, 9, 15, 16], then the closest 4 nodes are
// [8, 9, 15, 16].

// HeartBeat is sent to every peer nodes at "/heartbeat/receive".

// 2. For each HeartBeat, a node would randomly decide (this will change in Project 4)
// if it will create a new block. If so, add the block information into HeartBeatData and
// send the HeartBeatData to others.

// Every node has 32 peers, every peer will forward the new block to all peers according to PeerList.
// That is to make sure every user in the network would receive the new block. For this project.
// Every HeartBeatData takes 2 hops, which means after a node received a HeartBeatData from the original block maker,
// the remaining hop times is 1.
// ForwardHeartBeat will be call to do this
func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	Peers.Rebalance()
	peerMap := Peers.Copy()

	jsonObj, err := json.Marshal(heartBeatData)
	if err != nil {
		fmt.Println(err)
	}
	path := "/heartbeat/receive"
	for k, _ := range peerMap {
		fmt.Println("FORWARD/ !!!!!!!!!!!!!!!!!!!!!!!!!addr: ", k)
		url := k + path
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonObj))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(resp.Body)
		}
	}
}

// Start a while loop. Inside the loop, sleep for randomly 5~10 seconds,
// then use PrepareHeartBeatData() to create a HeartBeatData, and send it to all peers in the local PeerMap.
func StartHeartBeat() {
	for ifStarted {
		randTime := time.Duration(rand.Intn(6) + 5)
		fmt.Println("START/ Beating!! Time: ", randTime)
		time.Sleep(randTime * time.Second)
		peersJSON, err := Peers.PeerMapToJson()
		if err != nil {
			log.Panic(err)
		}
		fmt.Println(SBC)
		heartBeatData := data.PrepareHeartBeatData(&SBC, "", ID, peersJSON, SELF_ADDR)
		ForwardHeartBeat(heartBeatData)
	}
}

// Canonical(): This function prints the current canonical chain, and chains of all forks
// if there are forks. Note that all forks should end at the same height (otherwise there wouldn't be a fork).
// Example of the output of Canonical() function: You can have a different format, but it should be clean and clear.
func Canonical(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Not started Yet"))
		return
	}
	res := ""
	blocks := SBC.GetLatestBlocks()
	fmt.Println("Canonical////////////////////////// ", len(blocks))
	for _, v := range blocks {
		res += GetChain(v) + "\n"
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))
}

func Scene(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot read body"))
		return
	}

	var playData data.PlayData
	err = json.Unmarshal([]byte(body), &playData)
	fmt.Println()
	blocks := SBC.GetBlocks(playData.Height)

	if blocks == nil {
		fmt.Println("No block in that height")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot access block"))
		return
	}

	for i := 0; i < len(blocks); i++ {
		if playData.Height == 1 || accessVerify(playData.Id, blocks[i]) {
			SBC.AddPlayer(playData.Id, blocks[i])
			content, err := blocks[i].Value.Get("content")
			if err == nil {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(content))
				return
			}
		}
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Something wrong"))
}

func Rank(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot read body"))
		return
	}

	var blockinfo map[string]string
	err = json.Unmarshal([]byte(body), &blockinfo)
	hash := blockinfo["hash"]
	height, err := strconv.Atoi(blockinfo["height"])
	fmt.Println("rank: " + hash + " " + blockinfo["height"])
	if err == nil {
		block, notEmpty := SBC.GetBlock(int32(height), hash)
		if notEmpty {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(block.GetRankString()))
			return
		}
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Cannot access block"))
}

func accessVerify(id string, block p2.Block) bool {
	parentBlock := SBC.GetParentBlock(block)
	playersstr, err := parentBlock.Value.Get("playerlist")
	players := strings.Fields(playersstr)
	if err != nil {
		return false
	}
	for _, player := range players {
		if player == string(id) {
			return true
		}
	}
	return false
}

func Play(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot read body"))
		return
	}
	var playData data.PlayData
	err = json.Unmarshal([]byte(body), &playData)
	block, notEmpty := SBC.GetBlock(playData.Height, playData.Hash)
	fmt.Println(block)
	if notEmpty && reactVerify(playData.Id, block, playData.React) {
		secret := ""
		for i := 0; i < 16; i++ {
			secret += Hex[rand.Intn(16)]
		}
		SBC.AddCreator(playData.Id, secret, block)
		// send heartbeat data
		peersJSON, err := Peers.PeerMapToJson()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Cannot access peers"))
			return
		}
		heartBeatData := data.PrepareHeartBeatData(&SBC, "", ID, peersJSON, SELF_ADDR)
		heartBeatData.IfUpdateBlock = true
		heartBeatData.BlockJson = block.EncodeToJson()
		fmt.Println("The HeartBeat Data: ", heartBeatData.BlockJson)
		ForwardHeartBeat(heartBeatData)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(secret))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Something Wrong"))
}

func reactVerify(id string, block p2.Block, react string) bool {
	players := block.GetPlayer()
	fmt.Println("Players:")
	fmt.Println(players)
	for _, player := range players {
		if player == id {
			correct, err := block.Value.Get("react")
			if err == nil && correct == react {
				return true
			}
		}
	}
	return false
}

// Need information on creator id,
// The parent height and hash to create the block
// the information to create the block: game content and answer
func Create(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot read body"))
		return
	}

	var createinfo data.CreateData
	err = json.Unmarshal([]byte(body), &createinfo)
	fmt.Println("This is height: ")
	fmt.Println(createinfo.ParentHeight)
	block, notEmpty := SBC.GetBlock(createinfo.ParentHeight, createinfo.ParentHash)
	fmt.Println(block)
	if notEmpty && block.VerifySecret(createinfo.Id, createinfo.Secret) {
		CreateNewGameBlock(createinfo.ParentHash, createinfo.ParentHeight, createinfo.Id, createinfo.Content, createinfo.React, createinfo.Secret)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("create successfully"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Cannot create"))
}

func GetChain(block p2.Block) string {
	res := ""
	for true {
		blockJson := block.EncodeToJson()
		res += blockJson + "\n"
		if block.Header.Height == 0 || block.Header.Height == 1 {
			fmt.Println("empty")
			break
		}
		block = SBC.GetParentBlock(block)
		fmt.Println("parentBlock: ", block.Header.Height, "\n")
	}
	return res
}

// starts a new thread that tries different nonces to generate new blocks.
// Nonce is a string of 16 hexes such as "1f7b169c846f218a".
// Initialize the rand when you start a new node with something unique about each node,
// such as the current time or the port number. Here's the workflow of generating blocks:
func CreateNewGameBlock(parentHash string, parentHeight int32, creatorId string, content string, react string, secret string) {
	fmt.Println("CreatGame")
	if ifStarted {
		mpt := data.GenMPT(content, react)
		parentBlock, notEmpty := SBC.GetBlock(parentHeight, parentHash)
		if !notEmpty {
			return
		}
		var rank map[string]int32
		err := json.Unmarshal([]byte(parentBlock.GetRankString()), &rank)
		if rank[creatorId] == 0 {
			rank[creatorId] = 1
		} else {
			rank[creatorId] = rank[creatorId] + 1
		}
		block := SBC.GenBlock(mpt, rank, creatorId)
		peersJSON, err := Peers.PeerMapToJson()
		if err != nil {
			log.Panic(err)
		}
		heartBeatData := data.PrepareHeartBeatData(&SBC, creatorId, ID, peersJSON, SELF_ADDR)
		heartBeatData.IfNewBlock = true
		heartBeatData.BlockJson = block.EncodeToJson()
		heartBeatData.Secret = secret
		fmt.Println("The HeartBeat Data: ", heartBeatData.BlockJson)
		ForwardHeartBeat(heartBeatData)
	}
}

func Overview(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please start first"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot read body"))
		return
	}
	fmt.Println("here is OK")
	var playerinfo data.PlayData
	err = json.Unmarshal([]byte(body), &playerinfo)

	fmt.Println(playerinfo)

	res := SBC.GetOverview(playerinfo.Id)
	fmt.Println("This is res: " + res)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))
}
