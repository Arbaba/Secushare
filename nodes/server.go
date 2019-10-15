package nodes

import (
    "encoding/json"
    "net/http"
	"github.com/gorilla/mux"
    "github.com/Arbaba/Peerster/packets"
)

type Payload struct {
    Messages []packets.RumorMessage
    Peers []string
    PeerID string
}

func RunServer(gossiper *Gossiper) {

    r := mux.NewRouter()
  
	
    r.HandleFunc("/messages/recentList", func(w http.ResponseWriter, r *http.Request) {
       
        rumors := gossiper.GetLastRumorsFlush()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(Payload{Messages:rumors, PeerID: gossiper.Name})        

	})

	r.HandleFunc("/messages/send/{msg}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)

        if gossiper.SimpleMode {
            pkt := packets.GossipPacket{Simple: &packets.SimpleMessage{gossiper.Name, gossiper.RelayAddress(),vars["msg"]}}
            gossiper.SimpleBroadcast(pkt, gossiper.RelayAddress())
            gossiper.LogSimpleMessage(pkt.Simple)
        }else{
            pkt := packets.GossipPacket{Rumor: &packets.RumorMessage{gossiper.Name, gossiper.GetNextRumorID(gossiper.Name),vars["msg"]}}
            gossiper.RumorMonger(&pkt, gossiper.RelayAddress())
        }

        
	})
	
	

	
    r.HandleFunc("/peers/list", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(Payload{Peers: gossiper.Peers, PeerID: gossiper.Name})        
	})
	
	
    r.HandleFunc("/peers/add/{address}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        //check input
        gossiper.AddPeer(vars["address"])
        w.WriteHeader(http.StatusOK)

    })

    r.PathPrefix("/").Handler(http.FileServer(http.Dir("./nodes/static/")))


    http.ListenAndServe(":8080", r)
}