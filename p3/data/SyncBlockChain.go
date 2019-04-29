package data

import (
	"fmt"
	"sync"

	"../../p1"
	"../../p2"
)

// 1. SyncBlockChain aims to add a lock on the "BlockChain" data structure.
// Mostly it would be like "(1) lock (2) call BlockChain's function (3) unlock".
// 2. At this moment, there's no more detailed description of all functions.
// But you can ask questions about the functions which you get confused.

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: *p2.NewBlockChain()}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	blocks := sbc.bc.Get(height)
	if blocks == nil {
		return nil, false
	}
	return blocks, true
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	blocks, valid := sbc.Get(height)
	emptyBlock := p2.Block{}
	if !valid {
		return emptyBlock, false
	}
	for _, v := range blocks {
		if v.Header.ParentHash == hash {
			fmt.Println("GetBlock/ Find the Block!!!")
			return v, true
		}
	}
	return emptyBlock, false
}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(&block)
	sbc.mux.Unlock()
}

// CheckParentHash(): Yes. This function would check if the block with the given "parentHash"
// exists in the blockChain. If we have the parent block, we can insert the next block;
// if we don't have the parent block, we have to download the parent block before inserting the next block.
func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	if insertBlock.Header.ParentHash == "genesis" {
		return true
	}
	chain := sbc.bc.Chain
	for _, blocks := range chain {
		for _, block := range blocks {
			if block.Header.Hash == insertBlock.Header.ParentHash {
				return true
			}
		}
	}
	return false
}

//????
func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	block, err := p2.DecodeJsonToBlockChain(blockChainJson)
	if err != nil {
		sbc.mux.Unlock()
		return
	}

	sbc.bc = *block

	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	return sbc.bc.EncodeToJson()
}

// This function generates a new block after the current highest block.
// You may consider it "create the next block".
// For example, suppose we have blocks of height 1~5.
// GenBlock() would generate a new block of height 6, and its parentHash is the hash of the block at height 5.
func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie, rank map[string]int32, creatorId string) p2.Block {
	len := sbc.bc.Length
	fmt.Println("SBC length", len)
	parentHash := "genesis"
	if len > 0 {
		highestBlock := sbc.bc.Chain[len]
		parentHash = highestBlock[0].Header.Hash
	}
	// NewBlock(height int32, timeStamp int64, parentHash string, mpt p1.MerklePatriciaTrie,
	// 	rank map[string]int32, creator string, playerlist string, minorlist string)

	minorlist := map[string]string{}
	newBlock := p2.NewBlock(len+1, 1234567890, parentHash, mpt, rank, creatorId, "", minorlist)

	sbc.bc.Insert(newBlock)

	return *newBlock
}

func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}

func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	blocks := sbc.bc.GetLatestBlocks()
	sbc.mux.Unlock()
	return blocks
}

func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) p2.Block {
	sbc.mux.Lock()
	parentBlock := sbc.bc.GetParentBlock(block)
	sbc.mux.Unlock()
	return parentBlock
}

func (sbc *SyncBlockChain) GetBlocks(height int32) []p2.Block {
	sbc.mux.Lock()
	blocks := sbc.bc.GetBlocks(height)
	sbc.mux.Unlock()
	return blocks
}

//todo
func (sbc *SyncBlockChain) AddCreator(id int32, secret string, block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.AddCreator(string(id), secret, block.Header.Height, block.Header.Hash)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) UpdateBlock(block p2.Block, creator string) bool {
	sbc.mux.Lock()
	res := sbc.bc.UpdateBlock(block, creator)
	sbc.mux.Unlock()
	return res
}
