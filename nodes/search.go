
package nodes 


import (
	"github.com/Arbaba/Peerster/packets"
	"sync"
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


	return nil
}

//Searches for a file locally	
func SearchFile(keywords []string, filesInfo []FileMetaData, files map[string][]byte) []packets.SearchResult{


	return nil 
}
