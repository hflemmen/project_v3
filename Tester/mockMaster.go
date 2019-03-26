package main

import (
	"../orders/elevio/ordStruct"
	"../MakkerModul/connector"
	"../MakkerModul/decoding"
	"fmt"
	"os"
	"time"
	"math/rand"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

func main() {
	newOrder := make(chan ordStruct.ButtonEvent)
	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel(
		"handlerBackup.go", 44444, 55555)
	go func() {
		for {
			someOrder := ordStruct.ButtonEvent{Button:ordStruct.ButtonType(rand.Intn(2)),Floor:rand.Intn(3)}
			if (someOrder.Button == ordStruct.BT_HallDown && someOrder.Floor == 0) {
				someOrder.Button = ordStruct.BT_HallUp
			} else if (someOrder.Button == ordStruct.BT_HallUp && someOrder.Floor == 3) {
				someOrder.Button = ordStruct.BT_HallDown
			}
			time.Sleep(4 * time.Second)
			newOrder <- someOrder
		}
	}()
	fmt.Println("Started")
	for {
		select {

		case a := <-receiveLocal:
			fmt.Printf("Received from: %#v\n", a.Id)
			backupMsg := decoding.BackupMsg{Elevators: elevMap, Number: 1}
			// do something
			//key = cost
				msgChanLocal <- decoding.EncodeBackupMsg(backupMsg)
			msgChanLocal <- msg
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
