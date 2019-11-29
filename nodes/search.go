
package nodes 


import (
	"github.com/Arbaba/Peerster/packets"
	"sync"
	"math"
)

type CompletedSearches struct {
	List []packets.SearchReply
	sync.Mutex
}
//Queue where all searches spend at most 0.5 seconds
type SearchesQueue struct {
	Keywords[][]string
	sync.Mutex
}

func (queue *SearchesQueue) push(keywords []string){

}

func (queue *SearchesQueue) pop() []string{
	return nil
}

func (queue *SearchesQueue) isValid(keywords []string) bool{
	return true
}

func (completed *SearchesQueue) ProcessReplies(reply packets.SearchReply){
	//filter les replies
}

func (completed *CompletedSearches) Add(searchreply packets.SearchReply) {

}
 

//Returns a map with the correct budget for each selected peer
func ProcessBudget(budget uint64, peers []string)map[string]uint64{
	//no polymorphism...
	nbTargets := uint64(math.Min(float64(budget), float64(len(peers))))
	range_ := RandomRange(int(nbTargets), len(peers))
	budgets := make(map[string]uint64)
	for idx, peeridx := range range_{
		name := peers[peeridx]
		
		budgets[name] = budget / nbTargets
		if  budget % nbTargets != 0  && idx < int(budget % nbTargets){
			budgets[name] += 1
		}/*
		if idx == len(peers) && budget % nbTargets != 0 {
			if {
				budgets[name] = budget / nbTargets
			}
		}else if budget % nbTargets != 0{
			budgets[name] = budget / (nbTargets -1)
		}else { 
			budgets[name] = budget / nbTargets
		}*/
	}

	return budgets
}

//Searches for a file locally	
func SearchFile(keywords []string, filesInfo []FileMetaData, files map[string][]byte) []packets.SearchResult{


	return nil 
}
