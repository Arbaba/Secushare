package nodes

import (
	"Peerster/packets"
	"time"
)

func (gossiper *Gossiper) AntiEntropyLoop() {
	ticker := time.NewTicker(time.Second * time.Duration(gossiper.AntiEntropy))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			status := gossiper.GetStatusPacket()
			gossiper.SendPacketRandom(packets.GossipPacket{StatusPacket: status})
		}
	}

}
