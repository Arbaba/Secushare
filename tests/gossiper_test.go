package test

import (
	"Peerster/nodes"
	"Peerster/packets"
	"fmt"
	"sync"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}

	t.Fatal(fmt.Sprintf("%v != %v", a, b))
}
func TestGetRumor(t *testing.T) {
	rumors := make(map[string][]*packets.RumorMessage)
	gossiper := nodes.Gossiper{RumorsReceived: rumors}
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
	rumors := make(map[string][]*packets.RumorMessage)
	gossiper := nodes.Gossiper{RumorsReceived: rumors}
	assertEqual(t, gossiper.GetNextRumorID("A"), uint32(1))
}

func TestGetNextRumorIDAskTwo(t *testing.T) {
	rumor := packets.RumorMessage{Origin: "A", ID: 1, Text: "hello"}
	rumors := make(map[string][]*packets.RumorMessage)
	gossiper := nodes.Gossiper{RumorsReceived: rumors}
	gossiper.StoreRumor(packets.GossipPacket{Rumor: &rumor})
	assertEqual(t, gossiper.GetNextRumorID("A"), uint32(2))
}

func TestGetNextRumorID(t *testing.T) {
	rumors := make(map[string][]*packets.RumorMessage)
	gossiper := nodes.Gossiper{RumorsReceived: rumors}

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
	rumors := make(map[string][]*packets.RumorMessage)
	gossiper := nodes.Gossiper{RumorsReceived: rumors}
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
