
package nodes 


import (
	"github.com/Arbaba/Peerster/packets"
	"sync"
	"math"
	"time"
	"regexp"
	"fmt"
)

type CompletedSearches struct {
	List []packets.SearchReply
	sync.Mutex
}
//Queue where all searches spend at most 0.5 seconds
type SearchesQueue struct {
	searches []packets.SearchRequest
	sync.Mutex
}

func (queue *SearchesQueue) push(request packets.SearchRequest){
	queue.Lock()
	idx := len(queue.searches)
	queue.searches = append(queue.searches, request)
	queue.Unlock()
	ticker := time.NewTicker(time.Millisecond * time.Duration(500))
	defer ticker.Stop()
	select {
	case <-ticker.C:
		queue.Lock()
		queue.searches = append(queue.searches[idx:], queue.searches[idx + 1:]...)
		queue.Unlock()
	}
	
}

func (queue *SearchesQueue) pop(){
	queue.Lock()
	defer queue.Unlock()
	if len(queue.searches) == 0 {
		//should maybe return an error
		return
	}
	queue.searches = queue.searches[1:]
}

func (queue *SearchesQueue) isValid(request packets.SearchRequest) bool{
	queue.Lock()
	defer queue.Unlock()
	for _, searchRequest := range queue.searches {
		if searchRequest.Origin == request.Origin {
			for idx, kw := range request.Keywords {
				if searchRequest.Keywords[idx] == kw{
					return false
				}
			}
		}
	}
	return true
}

//returns the files matched by the origin of the reply
func ProcessReplies(reply packets.SearchReply, repliesHistory []packets.SearchReply) []string{
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

func (gossiper *Gossiper) SearchFile(keywords []string, tmpbudget *uint64,matchingPeers map[string][]string, resultsHistory []packets.SearchResults){
	var budget uint64
	if tmpbudget == nil {
		budget = uint64(2)
	}else {
		budget = *tmpbudget
	}

	budgets := 	ProcessBudget(budget - 1, gossiper.GetAllOrigins())
	for peerName, budget := range budgets {
		searchReq := &packets.SearchRequest{
			Origin: gossiper.Name, 
			Budget: budget,
			Keywords: keywords,
		}
		pkt := packets.GossipPacket{SearchRequest: searchReq}
		gossiper.SendDirect(pkt, peerName)
	}
	ticker := time.NewTicker(time.Second * time.Duration(1))
	
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if budget < 32 && tmpbudget == nil{
				newbudget := *tmpbudget *uint64(2)
				go gossiper.SearchFile(keywords, &newbudget, nbMatches)
				return
			}
		case results := <-	*gossiper.SearchChannel:
			/*
			Process search replies.
			Maintain a map filename -> (map chunk -> list origins)	
			*/
			if match {
				matchingPeers = append(matchingPeers, result.Origin )
				if len(nbMatches) == 2 {
					fmt.Println("SEARCH FINISHED")
					return
				}

			}


		}
	}

}

//Searches for a file locally	
func SearchFileLocally(keywords []string, filesInfo map[string]FileMetaData, files map[string][]byte) []packets.SearchResult{
	var results []packets.SearchResult
	for metafileHash, fileInfo := range filesInfo {
		for _, kw := range keywords {
			match, _ := regexp.MatchString(fmt.Sprintf("[[:alpha:]]%s[[:alpha:]]", kw), fileInfo.FileName )
			if match {
				var chunkMap []uint64
				for idx, chunkHash := range fileInfo.MetaFile {
					data, found := files[HexToString(chunkHash[:])]
					if found && len(data) > 0 {
						chunkMap = append(chunkMap, uint64(idx))
					}
				}
				searchResult := packets.SearchResult{
					FileName: fileInfo.FileName,
					MetaFileHash: []byte(metafileHash),
					ChunkMap: chunkMap,
					ChunkCount: uint64(len(fileInfo.MetaFile)),
				}
				results = append(results, searchResult)
				break
			}
			
		}
	
	}
	return results 
}