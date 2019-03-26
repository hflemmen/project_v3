package main

import (
	"./network/bcast"
	"./MakkerModul/connector"
	"./MakkerModul/decoding"
	//"./network/localip"
	"./network/idGenerator"
	"./network/peers"
	"./orders/elevio/ordStruct"
	"flag"
	"fmt"
	//"os"
	//"strconv"
	"strings"
	"time"
)

/*
	TODO
*/
type MsgFromHandlerToHandler struct {
	Id     string
	States ordStruct.Elevator
	Number int
}

func main() {
	time.Sleep(time.Millisecond) //kun så time er brukt
	elevMap := make(map[string]decoding.ElevatorStatus)
	var myName string
	flag.StringVar(&myName, "name", "", "id of this peer")
	flag.Parse()

	if myName == "" {
		myName = idGenerator.GetRandomID()
	}

	id := "Backup - " + myName
	fmt.Printf("#####################################\nStarting a backup with myName %v \n#####################################\n", myName)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	terminateTransmitter := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable, terminateTransmitter)
	go peers.Receiver(15647, peerUpdateCh)

	helloTx := make(chan MsgFromHandlerToHandler)
	helloRx := make(chan MsgFromHandlerToHandler)
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel(
		                          "Tester/mockMaster.go", 33333, 22222)

	// The example message. We just send one of these every second.
	/*go func() {
		helloMsg := MsgFromHandlerToHandler{Id: "MASTER", Number: 0}
		for {
			helloMsg.Number++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()*/

	for {
		select {
		case p := <-peerUpdateCh:
			var nonElevatorPeers []string
			for _, peer := range p.Peers {
				if strings.HasPrefix(peer, "Backup") || strings.HasPrefix(peer, "MASTER") {
					nonElevatorPeers = append(nonElevatorPeers, peer)
				}
			}
			fmt.Println("Non elevator peers:", nonElevatorPeers)
		sw:
			switch {
			case len(nonElevatorPeers) <= 1:
				if strings.HasPrefix(id, "MASTER") {
					id = "Backup - " + myName
					terminateTransmitter <- true
					go peers.Transmitter(15647, id, peerTxEnable, terminateTransmitter)
				}
			case true:
				for _, name := range nonElevatorPeers {
					if strings.HasPrefix(name, "MASTER") {
						fmt.Println("Glory to the master")
						break sw
					}
				}

				idOfMasterToBe := ""
				var tempCode string
				for _, name := range nonElevatorPeers {
					if strings.HasPrefix(name, "Backup - ") {
						tempCode = strings.Replace(name, "Backup - ", "", -1)
						if idOfMasterToBe < tempCode {
							idOfMasterToBe = tempCode
						}
					}
				}
				if strings.HasSuffix(id, idOfMasterToBe) && strings.HasPrefix(id, "Backup") {
					terminateTransmitter <- true
					fmt.Println("I am the master now")
					id = "MASTER - " + myName
					go peers.Transmitter(15647, id, peerTxEnable, terminateTransmitter)
				}
			}
			fmt.Println("This is the length og Peers - ", len(p.Peers))
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			fmt.Println("#####################################")

		case a := <-helloRx:
			if !strings.HasPrefix(a.Id,"Backup") {
				fmt.Println("HelloRx Id: ",a.Id)
				fmt.Printf("Received from: %#v\n", a.Id)
				elevMap[a.Id] = decoding.ElevatorStatus{E:a.States}
				latestOrder := a.States.CheckLatestOrder()
				if (latestOrder.Floor == -1) {
					break
				}
				fmt.Println("Latest Button: ",int(latestOrder.Button),"Latest Floor: ", latestOrder.Floor)
				backupMsg := decoding.BackupMsg{Elevators: elevMap,LatestOrder: latestOrder, Number: 1}
				msgChanLocal <- decoding.EncodeBackupMsg(backupMsg)
			}
		case a := <-receiveLocal:
			msg := decoding.DecodeElevatorMsg(a)
			fmt.Println("From Master to Elevators")
			helloTx <- MsgFromHandlerToHandler{Id: msg.E.ID}
	
		}
	}
}
