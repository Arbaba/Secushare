package nodes

import (
	"math/rand"
	"time"

	"github.com/Arbaba/Peerster/packets"
)

func (gossiper *Gossiper) UpdateRouting(origin, address string) {
	gossiper.RoutingTableMux.Lock()
	defer gossiper.RoutingTableMux.Unlock()
	gossiper.RoutingTable[origin] = address
}

func (gossiper *Gossiper) GetRandomRoute() (*packets.RumorMessage, string) {
	gossiper.RoutingTableMux.Lock()
	defer gossiper.RoutingTableMux.Unlock()
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	size := len(gossiper.RoutingTable)
	if size > 0 {
		i := r1.Intn(size)
		var origin, ip string
		for origin_, ip_ := range gossiper.RoutingTable {
			if i == 0 {
				origin = origin_
				ip = ip_
			}
			i -= 1
		}
		if len(origin) > 0 {
			rumor := &packets.RumorMessage{Origin: origin, ID: gossiper.GetNextRumorID(origin), Text: ""}
			return rumor, ip
		}
	}
	return nil, ""
}

func (gossiper *Gossiper) SendRandomRoute() {
	route, exceptip := gossiper.GetRandomRoute()
	if route != nil {
		gossiper.RumorMonger(&packets.GossipPacket{Rumor: route}, exceptip)
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
	origins := []string
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
	gossiper.PrivateMsgs[privatemsg.Origin] = append(gossiper.PrivateMsgs[privatemsg.Origin], privatemsg)
}


