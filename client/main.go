package main

import (
	"flag"
	"fmt"
	"net"
	"encoding/hex"
	"os"
	"github.com/Arbaba/Peerster/packets"
	"github.com/dedis/protobuf"
	"strings"
	"strconv"
)

func allEmpty( args...*string)bool{
	for _, s := range args {
		if *s !=  ""{
			return false
		}
	}
	return true
}


func allNonEmpty( args...*string)bool{
	for _, s := range args {
		if *s ==  ""{
			return false
		}
	}
	return true
}
func main() {
	uiport := flag.String("UIPort", "8080", "port for the UI client")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message")
	file := flag.String("file", "", "file to be indexed by the gossiper")
	request := flag.String("request", "", "Request of a chunk or metafile of this hash")
	keywords := flag.String("keywords", "", "Search keywords")
	budget := flag.String("budget", "", "Search budget")
	flag.Parse()
	target := fmt.Sprintf("127.0.0.1:%s", *uiport)
	simpleMsg := packets.Message{}
	if *uiport != ""{
		if *dest != "" && *msg != "" && allEmpty(file, request){
			simpleMsg.Destination = dest
			simpleMsg.Text = *msg

		}else if *file != "" && *dest != "" && *request != "" && allEmpty(msg){
			simpleMsg.Destination = dest

			simpleMsg.File = file
			b ,e:= hex.DecodeString(*request)
			if e != nil {
				fmt.Println("Unable to decoded hex hash")
				os.Exit(1)
			}
			simpleMsg.Request = &b
		}else if  *file != "" && allEmpty(dest, msg, request){
			simpleMsg.File = file

		} else if *msg != "" && allEmpty(dest, file, request){
			simpleMsg.Text = *msg

		}else if allNonEmpty(keywords, budget) && allEmpty(dest, file, request, msg){
			kw := strings.Split(*keywords, ",")
			b , err := strconv.ParseUint(*budget, 10, 64)
			if err  != nil {
				panic(err)

			}

			simpleMsg.Keywords = &kw
			simpleMsg.Budget = &b
		
		}else {
			fmt.Println("ERROR (Bad argument combination)")
			return
		}
	}else {
		fmt.Println("ERROR (Bad argument combination)")
		return
	}
	encodedPacket, err := protobuf.Encode(&simpleMsg)
	if err != nil {
		fmt.Println(err)
	}
	conn, err := net.Dial("udp", target)
	_, err = conn.Write(encodedPacket)
	if err != nil {
		fmt.Println("Error : ", err)
	}
}
