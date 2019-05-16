package p2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"../p1"
	"golang.org/x/crypto/sha3"
)

func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

type Header struct {
	Height int32 // index

	//The value must be in the UNIX timestamp format such as 1550013938
	TimeStamp int64

	// Block’s hash is the SHA3-256 encoded value of this string(note that you have to follow this specific order):
	// hash_str := string(b.Header.Height)
	// + string(b.Header.Timestamp) + b.Header.ParentHash + b.Value.Root + string(b.Header.Size)
	Hash string

	ParentHash string
	// The size is the length of the byte array of the block value
	Size int32

	// newly added
	// 	Find x such that y starts with at least 10 0's, while y is defined as
	// y = SHA-3(Hash of the parent block || x || root hash of MPT of the current block content)
	// where || is concatenation. Note that you will have to create an MPT of the block you wish to create before solving this puzzle.
	rank map[string]int32

	creator string

	playerList string

	minorList map[string]string
}

// Each block must have a value, which is a Merkle Patricia Trie.
// All the data are inserted in the MPT and then a block contains that MPT as the value. So the field definition is this:
// Value: mpt MerklePatriciaTrie
type Block struct {
	Header Header
	Value  p1.MerklePatriciaTrie
}

type BlockJson struct {
	Height     int32             `json:"height"`
	Timestamp  int64             `json:"timeStamp"`
	Hash       string            `json:"hash"`
	ParentHash string            `json:"parentHash"`
	Creator    string            `json:"creator"`
	Size       int32             `json:"size"`
	MPT        map[string]string `json:"mpt"`
	Rank       map[string]int32  `json:"rank"`
	playerList string            `json:"playerlist"`
	minorList  map[string]string `json:"minorlist"`
}

func calculateSize(mpt *p1.MerklePatriciaTrie) int32 {
	byteArray := []byte(fmt.Sprintf("%v", mpt))
	size := len(byteArray)
	return int32(size)
}

func NewBlock(height int32, timeStamp int64, parentHash string, mpt p1.MerklePatriciaTrie, rank map[string]int32, creator string, playerlist string, minorlist map[string]string) *Block {
	b := Block{}
	b.Initial(height, timeStamp, parentHash, mpt, rank, creator, playerlist, minorlist)
	return &b
}

func (b *Block) Initial(height int32, timeStamp int64, parentHash string, mpt p1.MerklePatriciaTrie, rank map[string]int32, creator string, playerlist string, minorlist map[string]string) {
	//create header
	size := calculateSize(&mpt)
	hash_str := string(height) + string(timeStamp) + parentHash + mpt.Get_root() + string(size)
	sum := sha3.Sum256([]byte(hash_str))
	hash := hex.EncodeToString(sum[:])

	//assign to block
	b.Header = Header{Height: height, TimeStamp: timeStamp, Hash: hash, ParentHash: parentHash, Size: size, rank: rank, creator: creator, playerList: playerlist, minorList: minorlist}
	b.Value = mpt
}

// takes a string that represents the JSON value of a block as an input, and decodes the input string back to a block instance.
// Note that you have to reconstruct an MPT from the JSON string,
// and use that MPT as the block's value.
// Argument: a string of JSON format
// Return value: a block instance

func DecodeFromJson(jsonString string) *Block {
	var result BlockJson
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		fmt.Println("some error deeper in bc")
		return nil
	} else {
		header := Header{Height: result.Height, TimeStamp: result.Timestamp, Hash: result.Hash, ParentHash: result.ParentHash, Size: result.Size, rank: result.Rank, creator: result.Creator, playerList: result.playerList, minorList: result.minorList}

		mpt := p1.MerklePatriciaTrie{}
		mpt.Initial()
		for key, value := range result.MPT {
			mpt.Insert(key, value)
		}
		b := Block{Header: header, Value: mpt}
		fmt.Println(header.ParentHash)
		return &b
	}
}

// Description: This function encodes a block instance into a JSON format string.
// Note that the block's value is an MPT, and you have to record all of the (key, value)
// pairs that have been inserted into the MPT in your JSON string. There's an example with details on Piazza.
// Here's a website that can encode and decode JSON string: Link (Links to an external site.)Links to an external site.
// Argument: a block or you may define this as a method of the block struct
// Return value: a string of JSON format

// Example of a block's JSON(decoded from JSON string):

// {
//     "hash":"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48",
//     "timeStamp":1234567890,
//     "height":1,
//     "parentHash":"genesis",
//     "size":1174,
//     "mpt":{
//         "charles":"ge",
//         "hello":"world"
//     }
// }

func (b *Block) EncodeToJson() string {
	str := "{"
	str += `"hash": "` + b.Header.Hash + `", `
	str += `"timeStamp": ` + fmt.Sprint(b.Header.TimeStamp) + `, `
	str += `"height": ` + fmt.Sprint(b.Header.Height) + `, `
	str += `"parentHash": "` + b.Header.ParentHash + `", `
	str += `"size": ` + fmt.Sprint(b.Header.Size) + `, `
	str += `"mpt": ` + b.Value.GetJsonString() + `, `
	str += `"creator": "` + b.Header.creator + `", `
	str += `"playerlist": "` + b.Header.playerList + `", `
	str += `"minorlist": ` + b.GetMinorString() + `, `
	str += `"rank": ` + b.GetRankString() + `}`
	return str
}

func (b *Block) GetMinorString() string {
	minorlist, err := json.Marshal(b.Header.minorList)
	if err != nil {
		return "{}"
	}
	return string(minorlist)
}

func (b *Block) GetRankString() string {
	fmt.Println("Rankstr : ")
	fmt.Println(b.Header.rank)
	rank, err := json.Marshal(b.Header.rank)
	if err != nil {
		return "{}"
	}
	return string(rank)
}

// ============================================================================================
// Each blockchain must contain two fields described below. Don't change the name or the data type.
// (1) Chain: map[int32][]Block
// This is a map which maps a block height to a list of blocks. The value is a list so that it can handle the forks.
// (2) Length: int32
// Length equals to the highest block height.

// Required functions:
// If arguments or return type is not specified, feel free to define them yourself.
// You may change the function's name, but make a comment to indicate which function you are implementing.
type BlockChain struct {
	Chain  map[int32][]Block
	Length int32 //最高height的
}

func NewBlockChain() *BlockChain {
	bc := BlockChain{Chain: make(map[int32][]Block), Length: 0}
	return &bc
}

// Description: This function takes a height as the argument, returns the list of blocks stored in that height or
// None if the height doesn't exist.
// Argument: int32
// Return type: []Block
func (bc *BlockChain) Get(height int32) []Block {
	if height > bc.Length {
		return nil
	}
	blocks := bc.Chain[height]
	if blocks == nil || len(blocks) == 0 {
		return nil
	}
	fmt.Println("Getting!!" + string(bc.Length))
	return blocks
}

// Description: This function takes a block as the argument,
// use its height to find the corresponding list in blockchain's Chain map.
// If the list has already contained that block's hash, ignore it because we don't store duplicate blocks;
// if not, insert the block into the list.
// Argument: block
func (bc *BlockChain) Insert(b *Block) {
	height := b.Header.Height
	blocks := bc.Chain[height]
	for i := 0; i < len(blocks); i++ {
		if b.Header.Hash == blocks[i].Header.Hash {
			return
		}
	}
	if bc.Chain[height] == nil {
		bc.Chain[height] = []Block{*b}
	} else {
		bc.Chain[height] = append(bc.Chain[height], *b)
	}

	if height > bc.Length {
		bc.Length = height
		fmt.Println("hahaha change the height")
	}
}

// Description: This function iterates over all the blocks, generate blocks'
// JsonString by the function you implemented previously, and return the list of those JsonStritgns.
// Return type: string

// Example of a blockchain's JSON:
// [
//     {
//         "hash":"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48",
//         "timeStamp":1234567890,
//         "height":1,
//         "parentHash":"genesis",
//         "size":1174,
//         "mpt":{
//             "hello":"world",
//             "charles":"ge"
//         }
//     },
//     {
//         "hash":"24cf2c336f02ccd526a03683b522bfca8c3c19aed8a1bed1bbc23c33cd8d1159",
//         "timeStamp":1234567890,
//         "height":2,
//         "parentHash":"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48",
//         "size":1231,
//         "mpt":{
//             "hello":"world",
//             "charles":"ge"
//         }
//     }
// ]

// type BlockChain struct {
// 	Chain  map[int32][]Block
// 	Length int32 //最高height的
// }

func (bc *BlockChain) EncodeToJson() (string, error) {
	ret := "["
	chain := bc.Chain
	for _, v := range chain {
		for i := 0; i < len(v); i++ {
			ret += v[i].EncodeToJson() + ","
		}
	}
	ret = ret[:len(ret)-1]
	ret += "]"
	return ret, nil
}

// Description: This function is called upon a blockchain instance.
// It takes a blockchain JSON string as input, decodes the JSON string back to a list of block JSON strings,
// decodes each block JSON string back to a block instance, and inserts every block into the blockchain.
// Argument: self, string
// Example of a blockchain's JSON:
// [
//     {
//         "hash":"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48",
//         "timeStamp":1234567890,
//         "height":1,
//         "parentHash":"genesis",
//         "size":1174,
//         "mpt":{
//             "hello":"world",
//             "charles":"ge"
//         }
//     },
//     {
//         "hash":"24cf2c336f02ccd526a03683b522bfca8c3c19aed8a1bed1bbc23c33cd8d1159",
//         "timeStamp":1234567890,
//         "height":2,
//         "parentHash":"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48",
//         "size":1231,
//         "mpt":{
//             "hello":"world",
//             "charles":"ge"
//         }
//     }
// ]

func (bc *BlockChain) DecodeFromJson(jsonString string) error {
	var result []BlockJson
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return err
	}

	if bc.Chain == nil {
		bc = NewBlockChain()
	}

	for i := 0; i < len(result); i++ {
		jsonObj, err := json.Marshal(result[i])
		if err != nil {
			return err
		}
		bc.Insert(DecodeFromJson(string(jsonObj)))
	}
	return nil
}

func DecodeJsonToBlockChain(jsonString string) (*BlockChain, error) {
	bc := NewBlockChain()
	bc.DecodeFromJson(jsonString)
	return bc, nil
}

// This function returns the list of blocks of height "BlockChain.length".
func (bc *BlockChain) GetLatestBlocks() []Block {
	return bc.Chain[bc.Length]
}

// This function takes a block as the parameter, and returns its parent block.
func (bc *BlockChain) GetParentBlock(block Block) Block {
	curHeight := block.Header.Height
	if curHeight <= 1 {
		return Block{Header: Header{}}
	}
	for _, parentBlock := range bc.Chain[curHeight-1] {
		if parentBlock.Header.Hash == block.Header.ParentHash {
			fmt.Println("FOUND PARENTBLOCK!!! ", parentBlock.Header.Hash, ": ", parentBlock.Header.Height)
			return parentBlock
		}
	}
	return Block{Header: Header{}}
}

func (bc *BlockChain) GetBlocks(height int32) []Block {
	if height > bc.Length || height < 1 {
		return nil
	}
	return bc.Chain[height]
}

func (bc *BlockChain) AddCreator(id string, secret string, height int32, hash string) {
	for k, v := range bc.Chain[height] {
		if v.Header.Hash == hash {
			bc.Chain[height][k].Header.minorList[id] = secret
			fmt.Println("Creator Added")
			fmt.Println(bc.Chain[height][k].Header.minorList)
		}
	}
}

func (bc *BlockChain) AddPlayer(id string, height int32, hash string) {
	for k, v := range bc.Chain[height] {
		if v.Header.Hash == hash {
			bc.Chain[height][k].Header.playerList += id + " "
			fmt.Println("hahaha " + bc.Chain[height][k].Header.playerList)
		}
	}
}

func (b *Block) VerifySecret(id string, secret string) bool {
	fmt.Println("Verifying: " + b.Header.minorList[id])
	if b.Header.minorList[id] == secret {
		return true
	}
	return false
}

func (b *Block) GetPlayer() []string {
	fmt.Println("list: " + b.Header.playerList)
	return strings.Fields(b.Header.playerList)
}

func (b *Block) GetMinor() map[string]string {
	return b.Header.minorList
}

func (bc *BlockChain) UpdateBlock(block Block, creator string) bool {
	for _, v := range bc.Chain[block.Header.Height] {
		if v.Header.Hash == block.Header.Hash && v.Header.creator == creator {
			//update player
			currentPlayer := v.GetPlayer()
			updatePlayer := block.GetPlayer()
			for _, playerId := range updatePlayer {
				found := false
				for _, currentId := range currentPlayer {
					if currentId == playerId {
						found = true
					}
				}
				if !found {
					v.Header.playerList += " " + playerId
				}
			}
			//update minor
			updateMinor := block.GetMinor()
			for id, secret := range updateMinor {
				v.Header.minorList[id] = secret
			}
			return true
		}
	}
	return false
}

// type BlockChain struct {
// 	Chain  map[int32][]Block
// 	Length int32 //最高height的
// }
func (bc *BlockChain) GetOverview(id string) string {
	res := ""
	passed := "No"
	for i := int32(1); i <= bc.Length; i++ {
		blocks := bc.Chain[i]
		fmt.Println(i)
		res += "LEVEL " + strconv.Itoa(int(i)) + ": \n"
		for j := 0; j < len(blocks); j++ {
			if blocks[j].Header.minorList[id] != "" {
				passed = "Yes"
			} else {
				passed = "No"
			}
			res += "Block " + string(j) + blocks[j].Header.Hash + "; Parent: " + blocks[j].Header.ParentHash + "; Passed: " + passed + "\n"
		}
		res += "======================================================================================================================================================\n"
	}
	return res
}
