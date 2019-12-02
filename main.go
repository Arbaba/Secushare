package main

import (
	"flag"
	"strings"

	"github.com/Arbaba/Peerster/nodes"
)

func main() {
	uiport, gossipAddr, name, peers, simpleMode, antiEntropy, guiPort, rtimer, networksize, stubbornTimeout, ex2, ex3, ex4, ackAll := parseCmd()
	gossiper := nodes.NewGossiper(*gossipAddr, *name, *uiport, peers, *simpleMode, *antiEntropy, *guiPort, *rtimer, *networksize, *stubbornTimeout, *ex2, *ex3, *ex4, *ackAll)
	//guiPort mandatory to run the webserver

	if *guiPort != "" {
		gossiper.LaunchGossiperGUI()
	} else {
		gossiper.LaunchGossiperCLI()
	}
}

func parseCmd() (*string, *string, *string, []string, *bool, *int64, *string, *int64, *int64, *int64, *bool, *bool, *bool, *bool) {
	//Parse arguments
	uiport := flag.String("UIPort", "8080", "port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	name := flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simpleMode := flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	antiEntropy := flag.Int64("antiEntropy", 10, "Use the given timeout in seconds for anti-entropy (relevant only for Part 2. If the flag is absent, the default anti-entropy duration is 10 seconds")
	guiPort := flag.String("GUIPort", "", "Port for the graphical interface")
	rtimer := flag.Int64("rtimer", 0, "Timeout in seconds for anti-entropy. If the flag is absent, the default anti-entropy duration is 10 seconds.")
	networksize := flag.Int64("N", 0, "Network sizes")
	stubbornTimeout := flag.Int64("stubbornTimeout", -1, "Stubborn timeout ")
	ex2 := flag.Bool("hw3ex2", false, "hw3ex2 flag")
	ex3 := flag.Bool("hw3ex3", false, "hw3ex3 flag")
	ex4 := flag.Bool("hw3ex4", false, "hw3ex4 flag")
	ackAll := flag.Bool("ackAll", false, "Flag to ack all TLC messages")
	flag.Parse()
	peersList := []string{}
	if *peers != "" {
		peersList = strings.Split(*peers, ",")
	}
	return uiport, gossipAddr, name, peersList, simpleMode, antiEntropy, guiPort, rtimer, networksize, stubbornTimeout, ex2, ex3, ex4, ackAll
}
