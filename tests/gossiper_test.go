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

func TestHighestRumor(t *testing.T) {
	rumors := make(map[string][]*packets.RumorMessage)
	gossiper := nodes.Gossiper{RumorsReceived: rumors}
	highestRumor := packets.RumorMessage{Origin: "A", ID: 10, Text: "hello"}
	gossiper.StoreRumor(packets.GossipPacket{Rumor: &highestRumor})
	assertEqual(t, gossiper.HighestRumor(highestRumor.Origin), &highestRumor)
	lowerRumor := packets.RumorMessage{Origin: "A", ID: 9, Text: "World"}
	gossiper.StoreRumor(packets.GossipPacket{Rumor: &lowerRumor})
	assertEqual(t, gossiper.HighestRumor(highestRumor.Origin), &highestRumor)

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
