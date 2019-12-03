package nodes

import (
	"github.com/Arbaba/Peerster/packets"
	"sync"
	"time"
)

const QSC_ROUNDS = 3

/*Stores the acks received per round*/
type AcksReceived struct {
	sync.Mutex
	store map[uint32][]string
}

func (acks *AcksReceived) Add(tlc *packets.TLCAck) {
	acks.Lock()
	defer acks.Unlock()
	if !Contains(acks.store[tlc.ID], tlc.Origin) {
		acks.store[tlc.ID] = append(acks.store[tlc.ID], tlc.Origin)
	}
}

func (acks *AcksReceived) Witnesses(id uint32) []string {
	acks.Lock()
	defer acks.Unlock()
	return acks.store[id]
}

func CreateAcksReceived() *AcksReceived {
	var v AcksReceived
	v.store = make(map[uint32][]string)
	return &v
}

//Keep track of other gossipers rounds and confirmedMessages
type RoundTable struct {
	sync.Mutex
	table map[string]int //Map gossiper name -> round
}

func CreateRoundTable() *RoundTable {
	var t RoundTable
	t.table = make(map[string]int)
	return &t
}

func (table *RoundTable) Increment(origin string) {
	table.Lock()
	defer table.Unlock()
	table.table[origin] += 1
}

func (table *RoundTable) GetRound(origin string) int {
	table.Lock()
	defer table.Unlock()
	return table.table[origin]
}

//Struct to keep track of the current round
type RoundState struct {
	sync.Mutex
	sentFirst         bool      //Indicates if the client sent a fist gossip with confirmation
	round             int       //Current node Round number
	majority          bool      //Tells whether the majority of confirmed messages has been collected during the current round
	sent              bool      //Tells wheter  the node has sent a message during the current round
	notify            *chan int //To notfy the gossiper when the round changes
	confirmedMessages [][]*packets.TLCMessage
}

func CreateRoundState(RoundNotify *chan int) *RoundState {
	var state RoundState
	state.notify = RoundNotify
	return &state
}
func (state *RoundState) RecordTLCMessage(tlc *packets.TLCMessage) {
	state.Lock()
	defer state.Unlock()
	if tlc.Confirmed != -1 {
		if len(state.confirmedMessages) <= state.round {
			var roundList []*packets.TLCMessage
			state.confirmedMessages = append(state.confirmedMessages, roundList)
		}
		state.confirmedMessages[state.round] = append(state.confirmedMessages[state.round], tlc)

	}
}

func (state *RoundState) RoundTLCMessages(round int) []*packets.TLCMessage {
	state.Lock()
	defer state.Unlock()
	//fmt.Println(state.confirmedMessages)
	return state.confirmedMessages[round]
}

func (state *RoundState) SetFirstSent() {
	state.Lock()
	defer state.Unlock()
	state.sentFirst = true
}
func (state *RoundState) HasSentFirst() bool {
	state.Lock()
	defer state.Unlock()
	return state.sentFirst
}

func (state *RoundState) SetMajority() {
	state.Lock()
	defer state.Unlock()
	state.majority = true
	state.advanceRound()

}
func (state *RoundState) GetRound() int {
	state.Lock()
	defer state.Unlock()
	return state.round

}

/*
func (state *RoundState) HasMajority() bool {
	state.Lock()
	defer state.Unlock()
	return state.majority
}*/

func (state *RoundState) SetSent() {
	state.Lock()
	defer state.Unlock()
	state.sent = true
	state.advanceRound()
}

func (state *RoundState) HasSent() bool {
	state.Lock()
	defer state.Unlock()
	return state.sent
}

/*
func (state *RoundState) IsRoundComplete() bool {
	state.Lock()
	defer state.Unlock()
	return state.majority && state.sent
}*/

func (state *RoundState) advanceRound() {
	//don't lock so it can be used in other methods
	if state.majority && state.sent {
		state.round += 1
		state.majority = false
		state.sent = false
		*state.notify <- state.round

	}
}
func (gossiper *Gossiper) Stubborn(tlcMessage *packets.TLCMessage) {
	if gossiper.StubbornTimeout <= 0 {
		return
	}
	ticker := time.NewTicker(time.Second * time.Duration(gossiper.StubbornTimeout))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if len(gossiper.AcksReceived.Witnesses(tlcMessage.ID)) < int(gossiper.NetworkSize/2) {
				pkt := packets.GossipPacket{TLCMessage: tlcMessage}
				gossiper.RumorMonger(&pkt, gossiper.Name)
			} else {
				return
			}
		}
	}
}

func (gossiper *Gossiper) ACKTLC(tlc *packets.TLCMessage) {

	tlcAck := packets.TLCAck{
		Origin:      gossiper.Name,
		ID:          tlc.ID,
		Destination: tlc.Origin,
		HopLimit:    gossiper.HOPLIMIT,
	}
	pkt := packets.GossipPacket{Ack: &tlcAck}
	gossiper.SendDirect(pkt, tlc.Origin)
	gossiper.LogSendTLCAck(&tlcAck, tlc.Origin)

}

func (gossiper *Gossiper) ProcessClientTLCMessages() {
	for {
		if !gossiper.Hw3ex2 || (!gossiper.RoundState.HasSentFirst() || !gossiper.RoundState.HasSent()) {
			select {
			case TLCMessage := <-gossiper.TLCBuffer:
				gossiper.RoundState.SetFirstSent() //modularize to avoid call each time
				gossiper.RoundState.SetSent()
				packet := packets.GossipPacket{TLCMessage: TLCMessage}
				gossiper.StoreLastPacket(packet)
				gossiper.StoreRumor(packet)
				gossiper.RumorMonger(&packet, gossiper.RelayAddress())
				go gossiper.Stubborn(TLCMessage)
			}
		}
	}
}

func (gossiper *Gossiper) ConfirmAndBroadcast(tlc packets.TLCMessage, witnesses []string) {
	tlc.Confirmed = int(tlc.ID)
	tlc.ID = gossiper.GetNextRumorID(tlc.Origin)
	tlc.VectorClock = gossiper.GetStatusPacket()
	pkt := &packets.GossipPacket{TLCMessage: &tlc}
	gossiper.RumorMonger(pkt, gossiper.RelayAddress())
	gossiper.LogRebroadcast(int(tlc.ID), witnesses)
}

func (gossiper *Gossiper) TrackRounds(notify *chan int) {
	var bestFit packets.TLCMessage
	for {
		select {
		case round := <-*notify:
			gossiper.LogAdvance(round)

			//Que sera consensus
			if gossiper.Hw3ex4 && gossiper.RoundState.GetRound()%QSC_ROUNDS == 0 && round > 0 {
				gossiper.Blockchain.Add(bestFit.TxBlock)
				gossiper.LogConsensus(bestFit.Origin, bestFit.ID)
			} else if gossiper.Hw3ex4 && gossiper.RoundState.GetRound()%QSC_ROUNDS == 1 {
				tlcs := gossiper.RoundState.RoundTLCMessages(round)

				for _, tlc := range tlcs {
					if tlc.Fitness > bestFit.Fitness {
						bestFit = *tlc
					}
				}
				pkt := packets.GossipPacket{
					TLCMessage: &bestFit,
				}
				gossiper.RumorMonger(&pkt, gossiper.RelayAddress())
			} else if gossiper.Hw3ex4 && gossiper.RoundState.GetRound()%QSC_ROUNDS == 2 {
				tlcs := gossiper.RoundState.RoundTLCMessages(round)

				for _, tlc := range tlcs {
					if tlc.Fitness > bestFit.Fitness {
						bestFit = *tlc
					}
				}
				pkt := packets.GossipPacket{
					TLCMessage: &bestFit,
				}
				gossiper.RumorMonger(&pkt, gossiper.RelayAddress())
			}
		}
	}
}
