package nodes

import (
    "encoding/json"
    "net/http"
    "regexp"
	"github.com/gorilla/mux"
    "github.com/Arbaba/Peerster/packets"
    "github.com/dedis/protobuf"
    "fmt"
    "strconv"
)

type Payload struct {
    Messages []packets.RumorMessage
    Peers []string
    PeerID string
}

func RunServer(gossiper *Gossiper) {

    r := mux.NewRouter()
  
	
    r.HandleFunc("/messages/recentList/{since}", func(w http.ResponseWriter, r *http.Request) {
       idx, err :=strconv.Atoi(mux.Vars(r)["since"])
       if err == nil {
            rumors := gossiper.GetLastRumorsSince(idx)
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(Payload{Messages:rumors, PeerID: gossiper.Name})        
       }
	})

	r.HandleFunc("/messages/send/{msg}", func(w http.ResponseWriter, r *http.Request) {
        
        vars := mux.Vars(r)
        msg:=vars["msg"]
        simpleMsg := packets.Message{msg}
        encodedPacket, err := protobuf.Encode(&simpleMsg)
        if err != nil {
            fmt.Println(err)
        }
        handleClient(gossiper, encodedPacket, len(encodedPacket))
        
        

        
	})
	
	

	
    r.HandleFunc("/peers/list", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(Payload{Peers: gossiper.Peers, PeerID: gossiper.Name})        
	})
	
	
    r.HandleFunc("/peers/add/{address}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        //check input
        match, _ := regexp.MatchString("([0-9]{1,3}.){3}[0-9]{1,3}", vars["address"])
        if match {
            gossiper.AddPeer(vars["address"])
        }
        w.WriteHeader(http.StatusOK)

    })

    r.PathPrefix("/").Handler(http.FileServer(http.Dir("./nodes/static/")))


    http.ListenAndServe(":8080", r)
}