package nodes

import (
	"time"

	"github.com/Arbaba/Peerster/packets"
)

func (gossiper *Gossiper) UpdateRouting(origin, address string) {
	gossiper.RoutingTableMux.Lock()
	defer gossiper.RoutingTableMux.Unlock()
	if origin != gossiper.Name {
		gossiper.RoutingTable[origin] = address
	}
}

func (gossiper *Gossiper) GetRouteRumor() *packets.RumorMessage {
	return &packets.RumorMessage{Origin: gossiper.Name, ID: gossiper.GetNextRumorID(gossiper.Name), Text: ""}

}

func (gossiper *Gossiper) SendRandomRoute() {
	route := gossiper.GetRouteRumor()
	if route != nil {
		pkt := &packets.GossipPacket{Rumor: route}
		gossiper.RumorMonger(pkt, gossiper.RelayAddress())
		gossiper.StoreLastPacket(*pkt)

		gossiper.StoreRumor(*pkt)

		gossiper.RumorMonger(pkt, gossiper.RelayAddress())
	}
}

func (gossiper *Gossiper) RouteRumorLoop() {
	ticker := time.NewTicker(time.Second * time.Duration(gossiper.Rtimer))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			gossiper.SendRandomRoute()
		}
	}

}

func (gossiper *Gossiper) GetRoute(origin string) string {
	gossiper.RoutingTableMux.Lock()
	defer gossiper.RoutingTableMux.Unlock()
	return gossiper.RoutingTable[origin]
}

func (gossiper *Gossiper) GetAllOrigins() []string {
	gossiper.RoutingTableMux.Lock()
	defer gossiper.RoutingTableMux.Unlock()
	var origins []string
	for origin, _ := range gossiper.RoutingTable {
		origins = append(origins, origin)
	}
	return origins
}

func (gossiper *Gossiper) SendPrivateMsg(privatemsg *packets.PrivateMessage) {
	gossiper.RoutingTableMux.Lock()
	defer gossiper.RoutingTableMux.Unlock()
	ip, found := gossiper.RoutingTable[privatemsg.Destination]
	if found {
		pkt := packets.GossipPacket{Private: privatemsg}
		gossiper.SendPacket(pkt, ip)
	}
}

func (gossiper *Gossiper) StorePrivateMsg(privatemsg *packets.PrivateMessage) {
	gossiper.PrivateMsgsMux.Lock()
	defer gossiper.PrivateMsgsMux.Unlock()
	if privatemsg.Origin == gossiper.Name {
		gossiper.PrivateMsgs[privatemsg.Destination] = append(gossiper.PrivateMsgs[privatemsg.Destination], privatemsg)
	} else {
		gossiper.PrivateMsgs[privatemsg.Origin] = append(gossiper.PrivateMsgs[privatemsg.Origin], privatemsg)
	}
}

func (gossiper *Gossiper) GetPrivateMsgs() map[string][]packets.PrivateMessage {
	gossiper.PrivateMsgsMux.Lock()
	defer gossiper.PrivateMsgsMux.Unlock()
	newMsgs := make(map[string][]packets.PrivateMessage)
	for origin, msgs := range gossiper.PrivateMsgs {
		for _, msg := range msgs {
			newMsgs[origin] = append(newMsgs[origin], *msg)
		}
	}
	return newMsgs
}
