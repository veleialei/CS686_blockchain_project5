# Project 5 Mid Point Specs


## Tasks Listed In Proposal 
This is the goal I listed in proposal for mid point milstone.

1. The information that will be produced in the first block: the basic game information.
2. The way the player can play with it, that they can follow each chain to experience the game.
3. The selection for the next block producer and verification methods.
4. The rank system

## Tasks I finished and how I finished
I almost finished all the tasks above, I may only modify them a little in furture. To accomplish those tasks, I did the following work: 
### 1. Added 2 files in p3/data folder: 
create.go which contain such a data structure
~~~~
type CreateData struct {
	Id           string `json:"id"`
	ParentHeight int32  `json:"parentHeight"`
	ParentHash   string `json:"parentHash"`
	Content      string `json:"hash"`
	React        string `json:"react"`
	Secret       string `json:"secret"`
}
~~~~
It contains the player id, new block' game content and parent block information, as well as the correct reaction for the given content.
The secret is the key the node send to player if they can create a block. They need to use this to verify they are the player but not anyone who knows the userid.

play.go which contain such a data structure
~~~~
type PlayData struct {
	Id     int32  `json:"id"`
	Addr   string `json:"addr"`
	Height int32  `json:"height"`
	Hash   string `json:"hash"`
	React  string `json:"react"`
}
~~~~
It contains the user id, information of the block it want to go and the player's reaction for the block


### 2. Edited some existing data structure
To make the block suitable for my project.
I remove Nonce and add 3 other elements in Header of a block for file p2/BlockChain.go:
~~~~
type Header struct {
  ...
	rank map[string]int32
	creator string
	playerList string
	minorList map[string]string
}
~~~~
I put the game information(the scene and reaction) into the MPT.

Previously the MPT is generated with some random content in p3/heartbeat.go.
Currently, I built some temporiary information inside it for quick testing. If the users' react not match with the MPT_A, they will failed to pass the game. 
~~~~
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
~~~~

### 3. Built 4 new API
/scene
A post request contains the json of PlayData (reaction part will be ignore)
To get the game content information

/play
A post request contains the json of PlayData
Once players' reaction is correct, they passed the game in the block, they will get their own create string for create block, and the node will send heartbeat to all peers, update the related block in them.

/create
A post request contains the json of CreateData
minor create new block, the node verify if they are in the minor list and have the correct secret. 
once created, send heartbeat to all peers, create the block in each node

/rank
A post request contains the PlayData (react will be ignore)
Then player can view creator rank of a certain block.

To support those API, I also edit SyncBlockChain.go and BlockChain.go. I add some verification method inside it and some getters for newly added data in Block struct.

## Play Process 

Through the work finished above, the basic game system is built. I will go through the play flow to present the play logic.

### If a user want to play the game
1. they need to visit /scene with the json format of PlayData. Once the node receive the PlayData, it begin the verifcation process: 1st, it should check if the block exists. 2nd, it should check if parentblock of the block contain the player's id, such that player cannot jump the block they must follow the chain if they haven't play the game. So the **task 2** is done.

If passed the verifcation, they would get the game information and be added into the PlayerList in the Block.Header. 

2. They can visit /play to post their reaction, if they got the scene.
The play API also need to verify the user information and block information to make sure the block exists and player is correct. However, unlike /scene, the block don't go to the parent block's playerlist but its own player list, as the player need to visit /scene first, then their data should be stored at the block. 
After verification, the node will check if the user's reaction match the value stored in MPT.
If the user successfully pass the game, their id will be stored at creater map in that block and get a secret key.

###  If a user want to create new game block
They can visit the /create with json format of CreateData.

The API will verify their right by first find the block then check if their id exist and match the secret key in the creater map.

If information is correct. the creater will create a block use current block as parent. The new block will have the new MPT with the formation creator upload, as well as new Creator and Rank map in Header. So the **task 3** is finished.

###  Other function: 
they can visit /rank with json format of PlayData
Player can see who create most block along the chain as long as they provide correct block information. The **task 4** of proposal also finished

## Conclusion
As stated above task 1 to 4 has finished, currently people can play, get rank and create new game block if verified. The task 1 is more abstract, which need some more explanation. For now, I provide two key value pairs to store the game information in MPT. As it is a block chain project, not a real game, it is more like a block chain game platform. Creator provide essential game data on chain and user's local game engine read the data, do visualization and all kinds of things. Player play the game on their own machine, if they pass the game ,the local game engine send the information to the block chain game platform. For now, I designed a very simple interaction way just like question and answer, it is easy to add more in MPT.


## Reference
I read through these articles for inspiration.
https://blockonomi.com/blockchain-games/
https://medium.com/crowdbotics/examples-of-blockchain-games-and-how-they-work-7fb0a1e76e2e
https://blockexplorer.com/news/best-blockchain-games/
https://www.blockchaingamer.biz/features/3283/most-anticipated-blockchain-games/
The first two articles explore the current technique of blockchain game and why they are suitable for block chain.
The last two of them lists some interesting game examples. 
