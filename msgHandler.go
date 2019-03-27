package main

import (
	"./MakkerModul/connector"
	"./MakkerModul/decoding"
	"./network/bcast"
	"./network/idGenerator"
	"./network/peers"
	"./orders/elevio/ordStruct"
	"flag"
	"fmt"
	"strings"
	"time"
)

///////	MSG HANDLER //////

const (
	PARTNER_NAME       = "elevator_fsm.go"
	LOCAL_RECEIVE_PORT = 55555
	LOCAL_SEND_PORT    = 44444
	BROAD_PORT         = 16569
	PEER_PORT          = 15647
)

type relationship int

const (
	Crashed      relationship = 2
	Connected                 = 1
	Disconnected              = 0
)

type pendingType int

const (
	FromMsgHandler pendingType = 2
	FromMaster                 = 1
	ToMaster                   = 0
)

type MsgHandler struct {
	MyElevStates ordStruct.Elevator
	MsgToElev    decoding.ElevatorMsg
	MsgFromElev  decoding.ElevatorMsg

	MsgToMaster   MsgFromHandlerToHandler
	MsgFromMaster MsgFromHandlerToHandler

	RelationElevator relationship
	RelationMaster   relationship
}

type MsgFromHandlerToHandler struct {
	Id     string
	States ordStruct.Elevator
	Number int
}

func main() {
	H := MsgHandler{MsgFromElev: decoding.ElevatorMsg{Number: 0},
		RelationElevator: Disconnected, RelationMaster: Disconnected}
	pendingUpdates := make(chan pendingType)
	var myName string

	flag.StringVar(&myName, "name", "", "id of this peer")
	flag.Parse()

	if myName == "" {
		myName = idGenerator.GetRandomID()
	}

	id := "Elev - " + myName
	fmt.Printf("#####################################\nelevtorMSGHANDLER with myName %v \n#####################################\n", myName)

	H.MsgToMaster.Id = id
	H.MsgToMaster.Number = 0

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(PEER_PORT, id, peerTxEnable, peerTxEnable)
	go peers.Receiver(PEER_PORT, peerUpdateCh)

	netTx := make(chan MsgFromHandlerToHandler)
	netRx := make(chan MsgFromHandlerToHandler)
	repeatTx := make(chan MsgFromHandlerToHandler)

	go bcast.Transmitter(BROAD_PORT, netTx)
	go bcast.Receiver(BROAD_PORT, netRx)

	go func() {
		for a := range repeatTx {
			for i := 0; i < 5; i++ {
				netTx <- a
				fmt.Println(a.Number)
			}
		}
	}()

	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel(
		PARTNER_NAME, LOCAL_RECEIVE_PORT, LOCAL_SEND_PORT)
	printStatus := make(chan bool)
	go func() {
		for {
			time.Sleep(time.Second)
			printStatus <- true
		}
	}()

	go func() {
		for update := range pendingUpdates {
			switch update {
			case FromMsgHandler:
				msgChanLocal <- decoding.EncodeElevatorMsg(H.MsgToElev)
			case FromMaster:
				msgChanLocal <- decoding.EncodeElevatorMsg(decoding.ElevatorMsg{E: H.MsgFromMaster.Elevators[Id]})
			case ToMaster:
				repeatTx <- H.MsgToMaster
			}
		}
	}()

	for {
		select {
		case p := <-peerUpdateCh:
			for _, name := range p.Lost {
				if strings.HasPrefix(name, "MASTER") {
					fmt.Println("LOST CONNECTION TO MASTER")
					H.RelationMaster = Disconnected
				}
			}
			if strings.HasPrefix(p.New, "MASTER") {
				fmt.Printf("Connected to master \n")
				H.RelationMaster = Connected
			}

		case a := <-netRx:
			if strings.HasPrefix(a.Id, "MASTER") {
				fmt.Printf("Received from (not local) %v\n", a.Id)
				H.MsgFromMaster.Elevators[Id].LightMatrix = a.LightsHall 
				H.MsgFromMaster.Elevators[Id] = a.Elevators[Id]
				pendingUpdates <- FromMaster
			}

		case a := <-receiveLocal:
			switch a {
			case "Connection lost":
				if H.MsgFromElev.Number > 0 {
					H.MsgToElev = H.MsgFromElev
				}
				H.RelationElevator = Crashed
			case "Connection established":
				if H.RelationElevator == Disconnected {
					H.RelationElevator = Connected
					switch H.RelationMaster {
					case Disconnected:
						pendingUpdates <- FromMsgHandler
					default:
						pendingUpdates <- FromMaster
					}
				}
			default:
				msg := decoding.DecodeElevatorMsg(a)
				if msg.Number > H.MsgFromElev.Number { // > ikke >= etter testing??
					H.MsgFromElev = msg
					if H.RelationMaster != Disconnected {
						H.MsgToMaster.States = msg.E
						H.MsgToMaster.Number++
						pendingUpdates <- ToMaster
					} else {
						H.MsgToElev = H.MsgFromElev
						H.MsgToElev.E.Order[0] = H.MsgToElev.E.LightMatrix[0]
						H.MsgToElev.E.Order[1] = H.MsgToElev.E.LightMatrix[1]
						pendingUpdates <- FromMsgHandler
					}
					H.MyElevStates = msg.E
				} else {
					H.RelationElevator = Crashed
					pendingUpdates <- FromMsgHandler
				}
				H.RelationElevator = Connected
			}
		case <-printStatus:
			//fmt.Printf("\nPRINTOUT %v\n\n", H.MyElevStates)
			/*default:
			if H.RelationElevator == Crashed {
				msgChanLocal <- decoding.EncodeMsg(H.MsgToElev)
				H.RelationElevator = PendingUpdates
			}
			//case msg from master
			*/
		}
	}
}
