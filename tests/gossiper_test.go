package test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Arbaba/Peerster/nodes"
	"github.com/Arbaba/Peerster/packets"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}

	t.Fatal(fmt.Sprintf("%v != %v", a, b))
}

/*
func initGossiper(){
	rumors := make(map[string][]*packets.RumorMessage)
	vectorClock := make(map[string]*packets.PeerStatus)
}*/
func getGossiper() *nodes.Gossiper {
	return nodes.NewGossiper("127.0.0.1:5000", "A", "12345", nil, false, 10, "8080", 10)
}
func TestGetRumor(t *testing.T) {
	gossiper := getGossiper()
	//gossiper := nodes.Gossiper{RumorsReceived: rumors}
	rumor := packets.RumorMessage{Origin: "A", ID: 10, Text: "hello"}
	lowerRumor := packets.RumorMessage{Origin: "A", ID: 9, Text: "World"}

	gossiper.StoreRumor(packets.GossipPacket{Rumor: &rumor})
	gossiper.StoreRumor(packets.GossipPacket{Rumor: &lowerRumor})

	assertEqual(t, gossiper.GetRumor(rumor.Origin, rumor.ID), &rumor)
	assertEqual(t, gossiper.GetRumor(rumor.Origin, lowerRumor.ID), &lowerRumor)

}

/*
func TestHighestRumor(t *testing.T) {
	rumors := make(map[string][]*packets.RumorMessage)
	gossiper := nodes.Gossiper{RumorsReceived: rumors}
	highestRumor := packets.RumorMessage{Origin: "A", ID: 10, Text: "hello"}
	gossiper.StoreRumor(packets.GossipPacket{Rumor: &highestRumor})
	assertEqual(t, gossiper.HighestRumor(highestRumor.Origin), &highestRumor)
	lowerRumor := packets.RumorMessage{Origin: "A", ID: 9, Text: "World"}
	gossiper.StoreRumor(packets.GossipPacket{Rumor: &lowerRumor})
	assertEqual(t, gossiper.HighestRumor(highestRumor.Origin), &highestRumor)

}*/

func TestGetNextRumorIDEmpty(t *testing.T) {
	gossiper := getGossiper()

	assertEqual(t, gossiper.GetNextRumorID("A"), uint32(1))
}

func TestGetNextRumorIDAskTwo(t *testing.T) {
	rumor := packets.RumorMessage{Origin: "A", ID: 1, Text: "hello"}
	gossiper := getGossiper()
	gossiper.StoreRumor(packets.GossipPacket{Rumor: &rumor})
	assertEqual(t, gossiper.GetNextRumorID("A"), uint32(2))
}

func TestGetNextRumorID(t *testing.T) {
	gossiper := getGossiper()

	//Init rumors
	for _, id := range []uint32{1, 2, 3, 5, 6, 7, 10} {
		rumor := packets.RumorMessage{Origin: "A", ID: id, Text: "hello"}
		gossiper.StoreRumor(packets.GossipPacket{Rumor: &rumor})
	}

	var askedIDs []uint32
	for i := 0; i < 4; i++ {
		nextID := gossiper.GetNextRumorID("A")
		askedIDs = append(askedIDs, nextID)
		rumor := packets.RumorMessage{Origin: "A", ID: nextID, Text: "hello"}
		gossiper.StoreRumor(packets.GossipPacket{Rumor: &rumor})
	}
	solution := []uint32{4, 8, 9, 11}
	for i, id := range askedIDs {
		assertEqual(t, id, solution[i])
	}
}

func TestRumorsOrdered(t *testing.T) {
	origins := []string{"A", "A", "B"}
	gossiper := getGossiper()
	var wg sync.WaitGroup
	wg.Add(len(origins))

	//Simulate a gossiper concurently receiving rumors from different nodes
	for nameidx, origin := range origins {
		go func(origin string, nameidx int) {
			defer wg.Done()

			for i := 0; i < 10; i++ {
				rumor := packets.RumorMessage{Origin: origin, ID: uint32(i + 10*nameidx), Text: fmt.Sprintf("Origin %s ID %d ", origin, i+10*nameidx)}
				gossiper.StoreRumor(packets.GossipPacket{Rumor: &rumor})

			}
		}(origin, nameidx)

	}

	wg.Wait()
	//Make sure rumors are stored in order
	for nameidx, origin := range origins {
		start := 0
		if nameidx == 1 && origin == "A" {
			start = 10
		}

		for idx, rumor := range gossiper.RumorsReceived[origin][start:] {
			if rumor.ID != uint32(idx+nameidx*10) {
				t.Fatal(fmt.Sprintf("Rumors from origin %s are not in order !", origin))
			}
		}
	}
}

func peerStatus(id string, nextID uint32) packets.PeerStatus {
	return packets.PeerStatus{Identifier: id, NextID: nextID}
}


func checkBudgets( budgets map[string]uint64,generalBudget, lowerBudget uint64) bool{
	lowerBudgetFound := false 
	for _, budget := range budgets {
		if budget != generalBudget && lowerBudgetFound && budget == lowerBudget{
			return false
		}else if budget == lowerBudget{
			lowerBudgetFound = true
		}else if budget != budget {
			return false
		}
	}

	return lowerBudgetFound
}
func TestProcessBudget(t *testing.T){
	budgets1 := nodes.ProcessBudget(2, []string{"A","B", "C","D","E"})
	
	assertEqual(t, checkBudgets(budgets1, 1,1), true)
	budgets2 := nodes.ProcessBudget(9, []string{"A","B", "C","D","E"})

	assertEqual(t, checkBudgets(budgets2, 2,1), true)
	fmt.Println(budgets1, budgets2)
	fmt.Println(nodes.ProcessBudget(12, []string{"A","B", "C","D","E"}))
}

func checkEqualArrays(t *testing.T, a, b []string){
	assertEqual(t, len(a), len(b))
	for idx, v := range a {
		assertEqual(t, v, b[idx])
	}
}
func TestMatches(t *testing.T){
	var matches nodes.Matches
	nodes.InitMatches(&matches)
	matches.AddMatch("profile.jpg", "A")
	matches.AddMatch("profile.jpg", "B")
	matches.AddMatch("video.mp4", "A")

	checkEqualArrays(t, matches.FindLocations("profile.jpg"), []string{"A","B"})
	checkEqualArrays(t, matches.FindLocations("video.mp4"), []string{"A"})


}
/*
func TestCompareStatusEqual(t *testing.T) {
	status := []packets.PeerStatus{peerStatus("A", 1), peerStatus("B", 2)}
	gossiper := nodes.Gossiper{RumorsReceived: make(map[string][]*packets.RumorMessage)}
	r,_,_ := gossiper.CompareStatus2(&packets.StatusPacket{status})
	assertEqual(t, r, 1)
}*/
/*
func TestCompareStatusDifferent(t *testing.T) {
	ackstatus := []packets.PeerStatus{peerStatus("B", 2), peerStatus("A", 1), peerStatus("D", 4)}

	status := []packets.PeerStatus{peerStatus("A", 2), peerStatus("B", 10), peerStatus("C", 100)}
	gossiper := nodes.Gossiper{RumorsReceived: make(map[string][]*packets.RumorMessage), PendingAcks: make(map[string][]packets.PeerStatus)}
	r := gossiper.CompareStatus(ackstatus, status)
	assertEqual(t, r[0], peerStatus("A", 1))
	assertEqual(t, r[1], peerStatus("B", 2))
	assertEqual(t, r[2], peerStatus("C", 1))
	assertEqual(t, len(r), 3)
}*/
