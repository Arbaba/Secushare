package main

import (
	"flag"
	"strings"
	"github.com/Arbaba/Peerster/nodes"
)

func main() {
	uiport, gossipAddr, name, peers, simpleMode, antiEntropy, guiPort := parseCmd()
	gossiper := nodes.NewGossiper(*gossipAddr, *name, *uiport, peers, *simpleMode, *antiEntropy, *guiPort)
	//guiPort mandatory to run the webserver
	if guiPort != nil {
		gossiper.LaunchGossiperGUI()
	}else {
		gossiper.LaunchGossiperCLI()
	}
}

func parseCmd() (*string, *string, *string, []string, *bool, *int64, *string) {
	//Parse arguments
	uiport := flag.String("UIPort", "8080", "port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	name := flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simpleMode := flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	antiEntropy := flag.Int64("antiEntropy", 10, "Use the given timeout in seconds for anti-entropy (relevant only for Part 2. If the flag is absent, the default anti-entropy duration is 10 seconds")
	guiPort := flag.String("GUIPort", "",  "Port for the graphical interface")
	flag.Parse()
	peersList := []string{}
	if *peers != "" {
		peersList = strings.Split(*peers, ",")
	}
	return uiport, gossipAddr, name, peersList, simpleMode, antiEntropy, guiPort
}




