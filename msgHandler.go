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
	"time"
)

type relationship int

const (
	Crashed        relationship = 2
	PendingUpdates              = 1
	UpToDate                    = 0
	Disconnected                = -1
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
	H := MsgHandler{MsgFromElev: decoding.ElevatorMsg{Number: -1},
		RelationElevator: Disconnected, RelationMaster: Disconnected}
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

	go peers.Transmitter(15647, id, peerTxEnable, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	helloTx := make(chan MsgFromHandlerToHandler)
	helloRx := make(chan MsgFromHandlerToHandler)
	repeatTx := make(chan MsgFromHandlerToHandler)
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	go func() {
		for a := range repeatTx {
			for i := 0; i < 5; i++ {
				helloTx <- a
				fmt.Println(a.Number)
			}
		}
	}()

	go func() {
		//helloMsg := MsgFromHandlerToHandler{Id: "I'm elevator msgHandler", Number: 0}
		for {
			//helloMsg.Number++
			//helloTx <- helloMsg
			time.Sleep(time.Second)
		}
	}()
	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel(
		"elevator_fsm.go", 55555, 44444)
	printStatus := make(chan bool)
	go func() {
		for {
			time.Sleep(time.Second)
			printStatus <- true
		}
	}()

	for {
		select {
		case p := <-peerUpdateCh:
			if p.New == "MASTER" {
				fmt.Printf("Connected to master \n")
				H.RelationMaster = PendingUpdates
			}
			for _, name := range p.Lost {
				if name == "MASTER" {
					fmt.Println("LOST CONNECTION TO MASTER")
					H.RelationMaster = Disconnected
				}
			}

		case a := <-helloRx:
			if a.Id == "MASTER" {
				fmt.Printf("Received from (not local) %v\n", a.Id)
				a.States.Order[0] = a.States.LightMatrix[0]
				a.States.Order[1] = a.States.LightMatrix[1]
				a.States.ID += " MsgHandler Connection"
				msgChanLocal <- decoding.EncodeElevatorMsg(decoding.ElevatorMsg{E: a.States})
			}

		case a := <-receiveLocal:
			switch a {
			case "Connection lost":
				H.MsgToElev = H.MsgFromElev
				H.RelationElevator = Crashed
			case "Connection established":
				if H.RelationElevator == Disconnected {
					H.RelationElevator = PendingUpdates
				}
			default:
				msg := decoding.DecodeElevatorMsg(a)
				msg.E.ID += " MsgHandler"
				if msg.Number >= H.MsgFromElev.Number { // > ikke >= etter testing??
					H.MsgFromElev = msg
					if H.RelationMaster != Disconnected {
						H.MsgToMaster.States = msg.E
						//helloTx <- H.MsgToMaster
						repeatTx <- H.MsgToMaster
						H.MsgToMaster.Number++
					} else {
						msg.E.Order[0] = msg.E.LightMatrix[0]
						msg.E.Order[1] = msg.E.LightMatrix[1]
						fmt.Printf("\nSent %v\n\n", msg)
						msg.E.ID += " MsgHandlerNoConnection"
						msgChanLocal <- decoding.EncodeElevatorMsg(msg)
					}
					H.MyElevStates = msg.E
				} else {
					H.RelationElevator = Crashed
				}
				H.RelationElevator = UpToDate
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
