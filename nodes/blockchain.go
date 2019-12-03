package nodes

import (
	"github.com/Arbaba/Peerster/packets"
	"sync"
)

type Blockchain struct {
	sync.Mutex
	blocks map[string]packets.BlockPublish
	head   packets.BlockPublish
}

func (blockchain *Blockchain) CheckValid(otherChain *packets.BlockPublish) bool {
	/*//parcourir block
	otherHead := otherChain.Transaction
	//
	founndmyChainHead := false
	for {
	}*/
	return false
}
