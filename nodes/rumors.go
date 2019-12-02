package nodes

import (
	"github.com/Arbaba/Peerster/packets"
	"sort"
)

//Returns the last packets as rumors (even for simpleMessages) and clears the list. Used to supply the GUI
func (gossiper *Gossiper) GetLastRumorsSince(idx int) []packets.RumorMessage {
	gossiper.LastPacketsMux.Lock()
	defer gossiper.LastPacketsMux.Unlock()
	//refactor
	var copy []packets.RumorMessage = nil
	if idx < len(gossiper.LastPackets) && len(gossiper.LastPackets) > 0 || idx == 0 {

		for _, packet := range gossiper.LastPackets[idx:] {
			if packet.Simple != nil {

				s := packet.Simple
				copy = append(copy, packets.RumorMessage{s.OriginalName, 0, s.Contents})
			} else if packet.Rumor != nil {
				copy = append(copy, *packet.Rumor)
			}

		}
	}
	return copy
}

//Stores the rumor in a map of list of ordered rumors. Each key of the map is a node orign
func (gossiper *Gossiper) StoreRumor(packet packets.GossipPacket) {
	var rumor packets.Rumorable
	if packet.Rumor != nil {
		rumor = packet.Rumor
	} else if packet.TLCMessage != nil {
		rumor = packet.TLCMessage
	}

	gossiper.rumorsMux.Lock()

	list := make([]packets.Rumorable, len(gossiper.RumorsReceived[rumor.GetOrigin()]))
	copy(list, gossiper.RumorsReceived[rumor.GetOrigin()])

	id := rumor.GetID()
	idx := sort.Search(len(list), func(i int) bool {
		return list[i].GetID() > id
	})
	if idx < len(list) {
		gossiper.RumorsReceived[rumor.GetOrigin()] = append(append(gossiper.RumorsReceived[rumor.GetOrigin()][:idx], rumor), list[idx:]...)
	} else {
		gossiper.RumorsReceived[rumor.GetOrigin()] = append(list, rumor)

	}
	gossiper.rumorsMux.Unlock()

	gossiper.UpdateVectorClock(rumor)

}

//Update the vector clock. Add Status or update nextID
func (gossiper *Gossiper) UpdateVectorClock(rumor packets.Rumorable) {
	gossiper.VectorClockMux.Lock()
	defer gossiper.VectorClockMux.Unlock()
	status, found := gossiper.VectorClock[rumor.GetOrigin()]
	if found && rumor.GetID() == status.NextID {
		status.NextID = rumor.GetID() + 1
		for {
			nextrumor := gossiper.GetRumor(rumor.GetOrigin(), status.NextID)
			if nextrumor != nil {
				status.NextID += 1
			} else {
				return
			}
		}
	} else if !found {

		nextID := uint32(1)
		if rumor.GetID() == nextID {
			nextID += 1
		}
		status = &packets.PeerStatus{Identifier: rumor.GetOrigin(), NextID: nextID}
		gossiper.VectorClock[rumor.GetOrigin()] = status
	}
}

//Returns the rumor
func (gossiper *Gossiper) GetRumor(origin string, id uint32) packets.Rumorable {
	gossiper.rumorsMux.Lock()
	defer gossiper.rumorsMux.Unlock()
	list := gossiper.RumorsReceived[origin]
	idx := sort.Search(len(list), func(i int) bool {
		return list[i].GetID() >= id
	})
	if idx < len(list) && list[idx].GetID() == id {
		return list[idx]
	}
	return nil

}

//Returns a GossipPacket with the rumor
func (gossiper *Gossiper) GetRumorPacket(origin string, id uint32) *packets.GossipPacket {
	rumorable := gossiper.GetRumor(origin, id)
	if rumorable != nil {
		if rumor, ok := rumorable.(*packets.RumorMessage); ok {
			return &packets.GossipPacket{Rumor: rumor}
		} else if tlcmsg, ok := rumorable.(*packets.TLCMessage); ok {
			return &packets.GossipPacket{TLCMessage: tlcmsg}
		}
	}
	return nil
}

func (gossiper *Gossiper) GetNextRumorID(origin string) uint32 {
	gossiper.rumorsMux.Lock()
	defer gossiper.rumorsMux.Unlock()
	rumors := gossiper.RumorsReceived[origin]
	if len(rumors) == 0 {
		return 1
	} else {
		prevID := uint32(0)
		for _, rumor := range rumors {
			if rumor.GetID() != uint32(prevID+1) {
				return prevID + 1
			}
			prevID += 1
		}
		return rumors[len(rumors)-1].GetID() + 1
	}
}
