package nodes

import (
	"github.com/Arbaba/Peerster/packets"
	"sync"
)

type Blockchain struct {
	sync.Mutex
	blocks map[string]packets.BlockPublish
	head   packets.BlockPublish
	names  []string
}

func CreateBlockchain() *Blockchain {
	var blockchain Blockchain
	blockchain.blocks = make(map[string]packets.BlockPublish)
	return &blockchain
}
func checkEquals(a, b []byte) bool {
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (blockchain *Blockchain) checkValidName(name string) bool {
	currentHash := (blockchain.head.PrevHash[:])
	var initHash [32]byte

	for checkEquals(currentHash, initHash[:]) {
		block, found := blockchain.blocks[HexToString(currentHash)]
		if !found {
			return false
		} else {
			if block.Transaction.Name == name {
				return false
			}
		}

		currentHash = (block.PrevHash[:])
	}
	return true
}

func (blockchain *Blockchain) CheckValid(otherChain *packets.BlockPublish) bool {

	blockchain.Lock()
	defer blockchain.Unlock()
	//Got issue with comparison, correct later
	if blockchain.checkValidName(otherChain.Transaction.Name) /* && (blockchain.head.Hash() == otherChain.PrevHash) */ {
		return true
	}
	return false
}

func (blockchain *Blockchain) Add(block packets.BlockPublish) {
	blockchain.Lock()
	defer blockchain.Unlock()
	h := block.Hash()
	blockchain.blocks[HexToString(h[:])] = block
	blockchain.head = block
	blockchain.names = append(blockchain.names, block.Transaction.Name)
}

func (blockchain *Blockchain) GetNames() []string {
	blockchain.Lock()
	defer blockchain.Unlock()
	return blockchain.names
}

func (blockchain *Blockchain) GetHeadHash() [32]byte {
	blockchain.Lock()
	defer blockchain.Unlock()
	return (&blockchain.head).Hash()
}

func (blockchain *Blockchain) GetHead() packets.BlockPublish {
	blockchain.Lock()
	defer blockchain.Unlock()
	return blockchain.head
}
