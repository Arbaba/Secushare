package nodes

import (
	"encoding/hex"
	"encoding/json"

	"fmt"
	"github.com/Arbaba/Peerster/packets"
	"github.com/dedis/protobuf"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type LogsContainer struct {
	sync.Mutex
	logs    []string
	counter int
}

func (l *LogsContainer) Add(s string) {
	l.Lock()
	defer l.Unlock()
	l.logs = append(l.logs, s)
}

func (l *LogsContainer) Flush() []string {
	l.Lock()
	defer l.Unlock()
	buf := l.logs[l.counter:]
	l.counter = len(l.logs)
	return buf
}

type Payload struct {
	Messages        []packets.RumorMessage
	Peers           []string
	PeerID          string
	Origins         []string
	PrivateMessages map[string][]packets.PrivateMessage
	Files           []FilePayload
	Logs            []string
	Matches         []string
}

type FilePayload struct {
	FileName string
	FileSize uint32
	MetaHash string
}

type PrivateMessageRequest struct {
	PrivateMessages map[string]int
	PeerID          string
}

func RunServer(gossiper *Gossiper) {

	r := mux.NewRouter()
	r.HandleFunc("/logs/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Payload{Logs: gossiper.LogsContainer.Flush(), PeerID: gossiper.Name})

	})

	r.HandleFunc("/messages/recentList/{since}", func(w http.ResponseWriter, r *http.Request) {
		idx, err := strconv.Atoi(mux.Vars(r)["since"])
		if err == nil {
			rumors := gossiper.GetLastRumorsSince(idx)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Payload{Messages: rumors, PeerID: gossiper.Name})
		}
	})

	r.HandleFunc("/messages/private/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msgs := gossiper.GetPrivateMsgs()
		json.NewEncoder(w).Encode(Payload{PrivateMessages: msgs, PeerID: gossiper.Name})
	})

	r.HandleFunc("/messages/send/{msg}", func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		msg := vars["msg"]
		simpleMsg := packets.Message{Text: msg}
		encodedPacket, err := protobuf.Encode(&simpleMsg)
		if err != nil {
			fmt.Println(err)
		}
		handleClient(gossiper, encodedPacket, len(encodedPacket))

	})

	r.HandleFunc("/messages/private/send/{dest}/{msg}", func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		msg := vars["msg"]
		dest := vars["dest"]

		simpleMsg := packets.Message{Text: msg, Destination: &dest}
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

	r.HandleFunc("/routing/origins", func(w http.ResponseWriter, r *http.Request) {
		//check input
		w.Header().Set("Content-Type", "application/json")
		payload := Payload{PeerID: gossiper.Name, Origins: gossiper.GetAllOrigins()}
		json.NewEncoder(w).Encode(payload)

	})

	r.HandleFunc("/routing/origins", func(w http.ResponseWriter, r *http.Request) {
		//check input
		w.Header().Set("Content-Type", "application/json")
		payload := Payload{PeerID: gossiper.Name, Origins: gossiper.GetAllOrigins()}
		json.NewEncoder(w).Encode(payload)

	})

	r.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		file, handler, err := r.FormFile("newFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")

			fmt.Println(err)
			return
		}
		defer file.Close()
		msg := packets.Message{File: &handler.Filename}
		encodedPacket, err := protobuf.Encode(&msg)
		handleClient(gossiper, encodedPacket, len(encodedPacket))

	})

	r.HandleFunc("/files/list", func(w http.ResponseWriter, r *http.Request) {
		//check input
		w.Header().Set("Content-Type", "application/json")
		var files []FilePayload
		gossiper.FilesInfoMux.Lock()
		for k, fileInfo := range gossiper.FilesInfo {
			files = append(files, FilePayload{fileInfo.FileName, fileInfo.FileSize, k})
		}
		gossiper.FilesInfoMux.Unlock()
		payload := Payload{PeerID: gossiper.Name, Files: files}
		json.NewEncoder(w).Encode(payload)

	})

	r.HandleFunc("/files/download", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		hash := r.FormValue("request")
		dest := r.FormValue("destination")
		name := r.FormValue("filename")
		h, e := hex.DecodeString(hash)
		if e != nil {
			fmt.Println(e)
			return
		}
		msg := packets.Message{Destination: &dest, Request: &h, File: &name}
		encodedPacket, err := protobuf.Encode(&msg)
		if err != nil {
			fmt.Println(err)
		}
		handleClient(gossiper, encodedPacket, len(encodedPacket))
	})
	r.HandleFunc("/files/download/{filename}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		vars := mux.Vars(r)
		filename := vars["filename"]
		gossiper.FilesInfoMux.Lock()
		for h, fileInfo := range gossiper.FilesInfo {
			if fileInfo.FileName == filename {
				gossiper.FilesInfoMux.Unlock()

				metahash, e := hex.DecodeString(h)
				if e != nil {
					return
				}
				msg := packets.Message{File: &filename, Request: &metahash}
				encodedPacket, err := protobuf.Encode(&msg)
				if err != nil {
					fmt.Println(err)
				}
				handleClient(gossiper, encodedPacket, len(encodedPacket))
				return
			}
		}
		gossiper.FilesInfoMux.Unlock()

	})

	r.HandleFunc("/files/search/{keywords}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keywords := strings.Split(vars["keywords"], ",")
		msg := packets.Message{Keywords: &keywords}
		encodedPacket, err := protobuf.Encode(&msg)
		if err != nil {
			fmt.Println(err)
		}
		handleClient(gossiper, encodedPacket, len(encodedPacket))

	})

	r.HandleFunc("/files/matches", func(w http.ResponseWriter, r *http.Request) {
		var files []string
		gossiper.Matches.Lock()
		for fname, _ := range gossiper.Matches.Locations {
			files = append(files, fname)
		}
		gossiper.Matches.Unlock()
		payload := Payload{PeerID: gossiper.Name, Matches: files}
		json.NewEncoder(w).Encode(payload)

	})

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./nodes/static/")))

	http.ListenAndServe(":"+gossiper.GUIPort, r)
}
