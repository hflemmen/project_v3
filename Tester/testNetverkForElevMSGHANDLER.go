package main

import (
	"../network/bcast"
	"../network/localip"
	"../network/peers"
	"../orders/elevio/ordStruct"
	"flag"
	"fmt"
	"os"
	"time"
	//"math/rand"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type MsgFromHandlerToHandler struct {
	Id string
	States ordStruct.Elevator
	Number int
}

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "MASTER", "id of this peer")
	flag.Parse()
	e := ordStruct.ElevatorInit("Hei",4)
	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable,peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan MsgFromHandlerToHandler)
	helloRx := make(chan MsgFromHandlerToHandler)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)
	newOrder := make(chan ordStruct.ButtonEvent)
	// The example message. We just send one of these every second.
	go func() {
		someOrder := ordStruct.ButtonEvent{Button:ordStruct.BT_HallUp,Floor:0}
		for {
			time.Sleep(2 * time.Second)
			newOrder <- someOrder
		}
	}()
	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			if a.Id != "MASTER"{
				fmt.Printf("Received from: %#v\n", a.Id)
				a.States.ID += " Master"
				e = a.States
				msg := MsgFromHandlerToHandler{Id: "MASTER", States:a.States, Number : 1}
				helloTx <- msg
			}
		case a := <-newOrder:
			if a.Floor != e.Floor {
				e.LightMatrix[int(a.Button)][a.Floor] = 1
			} else {
				e.LightMatrix[int(a.Button)][a.Floor] = 0
			}
			msg := MsgFromHandlerToHandler{Id: "MASTER", States:e,Number : 1}
			helloTx <- msg
		}
	}
}
