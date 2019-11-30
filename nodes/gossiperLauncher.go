package nodes

/*
This file serves to handle the high level behavior of the gossiper

*/
import (
	"fmt"
	"net"

	"github.com/Arbaba/Peerster/packets"
	"github.com/dedis/protobuf"
)

func (gossiper *Gossiper) LaunchGossiperCLI() {
	go listenClient(gossiper)
	go listenGossip(gossiper)
	if gossiper.Rtimer > 0 {
		go gossiper.SendRandomRoute()
		go gossiper.RouteRumorLoop()
	}
	gossiper.AntiEntropyLoop()
}

func (gossiper *Gossiper) LaunchGossiperGUI() {
	go listenClient(gossiper)
	go listenGossip(gossiper)
	go gossiper.AntiEntropyLoop()
	if gossiper.Rtimer > 0 {
		go gossiper.SendRandomRoute()
		go gossiper.RouteRumorLoop()
	}
	RunServer(gossiper)

}

func UdpConnection(address string) (*net.UDPAddr, *net.UDPConn) {
	udpAddr, _ := net.ResolveUDPAddr("udp4", address)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	return udpAddr, udpConn
}

func listenClient(gossiper *Gossiper) {
	conn := gossiper.ClientConn
	defer conn.Close()
	for {
		message := make([]byte, 1 << 13)
		rlen, _, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}
		go handleClient(gossiper, message[:rlen], rlen)

	}
}
func listenGossip(gossiper *Gossiper) {
	conn := gossiper.GossipConn
	defer conn.Close()
	for {
		message := make([]byte, 1 << 14)
		rlen, raddr, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}
		go handleGossip(gossiper, message, rlen, raddr)

	}
}

//Gossiper behavior on reception on a client message
func handleClient(gossiper *Gossiper, message []byte, rlen int) {
	var msg packets.Message
	protobuf.Decode(message[:rlen], &msg)
	gossiper.LogClientMsg(msg)

	if gossiper.SimpleMode {
		packet := packets.GossipPacket{
			Simple: &packets.SimpleMessage{
				OriginalName:  gossiper.Name,
				RelayPeerAddr: gossiper.RelayAddress(),
				Contents:      msg.Text,
			}}
		sourceAddress := packet.Simple.RelayPeerAddr
		packet.Simple.RelayPeerAddr = gossiper.RelayAddress()
		packet.Simple.OriginalName = gossiper.Name
		gossiper.StoreLastPacket(packet)
		gossiper.SimpleBroadcast(packet, sourceAddress)
	} else {
		//TODO: Refactor with AllNonEmpty
		if msg.Destination != nil && msg.File != nil && msg.Request != nil {
			dataReply, filemetadata := gossiper.DownloadMetaFile(HexToString(*msg.Request), *msg.Destination, *msg.File)
			gossiper.DownloadFile(dataReply, filemetadata)
		}else if msg.File != nil && msg.Request != nil {
			//TODO: check request
			gossiper.DownloadFoundFile(*msg.File)
		} else if msg.Destination != nil {

			privatemsg := &packets.PrivateMessage{
				Origin:      gossiper.Name,
				ID:          0,
				Text:        msg.Text,
				Destination: *msg.Destination,
				HopLimit:    gossiper.HOPLIMIT,
			}
			gossiper.SendPrivateMsg(privatemsg)
			gossiper.StorePrivateMsg(privatemsg)
		} else if msg.File != nil {
			gossiper.ScanFile(*msg.File)

		}else if msg.Keywords != nil {
			if msg.Budget != nil{
				gossiper.SearchFile(*msg.Keywords, *msg.Budget , make(map[string][]string), false)
			}else {
				gossiper.SearchFile(*msg.Keywords, uint64(2), make(map[string][]string), true)
			}
		} else {
			//RumorMongering
			packet := packets.GossipPacket{
				Rumor: &packets.RumorMessage{
					Origin: gossiper.Name,
					ID:     gossiper.GetNextRumorID(gossiper.Name),
					Text:   msg.Text},
			}
			gossiper.StoreLastPacket(packet)

			gossiper.StoreRumor(packet)

			gossiper.RumorMonger(&packet, gossiper.RelayAddress())
		}

	}
	gossiper.LogPeers()
}

//Gossiper behavior on reception of another node packet
func handleGossip(gossiper *Gossiper, message []byte, rlen int, raddr *net.UDPAddr) {
	var packet packets.GossipPacket
	protobuf.Decode(message[:rlen], &packet)
	peerAddr := fmt.Sprintf("%s:%d", raddr.IP, raddr.Port)
	if packet.Simple != nil {
		gossiper.AddPeer(packet.Simple.RelayPeerAddr)
		gossiper.LogPeers()
		gossiper.LogSimpleMessage(packet.Simple)

		sourceAddress := packet.Simple.RelayPeerAddr
		packet.Simple.RelayPeerAddr = gossiper.RelayAddress()
		gossiper.SimpleBroadcast(packet, sourceAddress)
		gossiper.StoreLastPacket(packet)

	} else if rumor := packet.Rumor; packet.Rumor != nil {

		gossiper.LogRumor(rumor, peerAddr)
		gossiper.AddPeer(peerAddr)
		gossiper.LogPeers()
		gossiper.UpdateRouting(rumor.Origin, peerAddr, rumor.ID)
		gossiper.LogDSDVRumor(rumor, peerAddr)

		rumor := gossiper.GetRumor(rumor.Origin, rumor.ID)
		if rumor != nil {
			//Rumor was already received, hence we discard the packet
			return
		}
		gossiper.StoreLastPacket(packet)

		gossiper.StoreRumor(packet)
		gossiper.SendPacket(packets.GossipPacket{StatusPacket: gossiper.GetStatusPacket()}, peerAddr)
		gossiper.RumorMonger(&packet, peerAddr)

	} else if packet.StatusPacket != nil {
		gossiper.LogStatusPacket(packet.StatusPacket, peerAddr)
		gossiper.AddPeer(peerAddr)
		gossiper.LogPeers()
		for _, status := range packet.StatusPacket.Want {
			//generate and identifier to retrieve to correct ACK channel
			//A goroutine with an ackchannel is created for each rumor sent.
			//hence we pass the status via the correct channel (see rumormongering.go)
			ackID := gossiper.AckID(status.Identifier, status.NextID, peerAddr)
			gossiper.AcksChannelsMux.Lock()
			ackChannel, waitingForIt := gossiper.AcksChannels[ackID]
			gossiper.AcksChannelsMux.Unlock()
			if waitingForIt {
				*ackChannel <- packet.StatusPacket
			} else {
				gossiper.AckRandomStatusPkt(packet.StatusPacket, peerAddr)
			}
		}

	} else if private := packet.Private; packet.Private != nil {
		gossiper.UpdateRouting(private.Origin, peerAddr, private.ID)
		gossiper.LogDSDVPrivate(private, peerAddr)

		if private.Destination == gossiper.Name {
			gossiper.StorePrivateMsg(private)
			gossiper.LogPrivateMsg(private)
		} else if private.HopLimit > 0 {
			private.HopLimit -= 1
			gossiper.SendPrivateMsg(private)
		}
	} else if reply := packet.DataReply; packet.DataReply != nil {
		if reply.Destination != gossiper.Name && reply.HopLimit > 0 {
			reply.HopLimit -= 1
			gossiper.SendDirect(packet, reply.Destination)
		}else {
			gossiper.DataBufferMux.Lock()
 			channel, found := gossiper.DataBuffer[HexToString(packet.DataReply.HashValue)]
			gossiper.DataBufferMux.Unlock()
			if found {
				
				*channel <- *reply
			}	
		}
		
	} else if request := packet.DataRequest; request != nil {
		//fmt.Println(request.Origin, request.Destination, request.HopLimit)
		if request.Destination != gossiper.Name && request.HopLimit > 0 {
			request.HopLimit -= 1
			gossiper.SendDirect(packet, request.Destination)
		} else if request.Destination == gossiper.Name{

			reply := packets.DataReply{Origin: gossiper.Name,
				Destination: request.Origin,
				HopLimit:    gossiper.HOPLIMIT,
			}
			reply.HashValue = []byte(request.HashValue)
			pkt := packets.GossipPacket{DataReply: &reply}


			gossiper.FilesInfoMux.Lock()
			hashString := HexToString(request.HashValue)
			filemetadata, foundmetadata := gossiper.FilesInfo[hashString]
			gossiper.FilesInfoMux.Unlock()
			if foundmetadata {
				for _, chunkHash := range filemetadata.MetaFile {
					reply.Data = append(reply.Data, chunkHash[:]...)
				}

			} else {
				gossiper.FilesMux.Lock()
				chunkData, foundChunk := gossiper.Files[hashString]
				gossiper.FilesMux.Unlock()
				if foundChunk {
					reply.Data = chunkData
				}

			}
			gossiper.SendDirect(pkt, reply.Destination)
			

		}

	}else if searchRequest := packet.SearchRequest; searchRequest != nil {
		gossiper.UpdateRouting(searchRequest.Origin, peerAddr, 0)
		reply := gossiper.SearchFilesLocally(searchRequest)
		if len(reply.Results) > 0 {
			pkt := packets.GossipPacket{SearchReply: &reply}
			gossiper.SendDirect(pkt, searchRequest.Origin)
		}
		fmt.Println(reply.Results)
		gossiper.ForwardSearchRequest(searchRequest)
	}else if searchReply:= packet.SearchReply; searchReply != nil {
		fmt.Println(*searchReply)
		gossiper.UpdateRouting(searchReply.Origin, peerAddr, 0)
		if searchReply.Destination == gossiper.Name{
			gossiper.SearchChannel <- *searchReply
		}else if searchReply.HopLimit > 0{
			searchReply.HopLimit -= 1
			gossiper.SendDirect(packet, searchReply.Destination)
		}
		
	}
}
