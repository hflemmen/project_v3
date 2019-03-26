package main

import (
	//"../orders/elevio/ordStruct"
	"../MakkerModul/connector"
	"../MakkerModul/decoding"
	"../cost"
	"fmt"
	//"os"
	//"time"
	//"math/rand"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

func main() {
	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel("../handlerBackup.go", 22222, 33333)
	/*
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
	*/
	fmt.Println("Started")
	for {
		select {

		case a := <-receiveLocal:
			fmt.Println("New OrderUpdate Master")
			msg := decoding.DecodeBackupMsg(a)
			elevId := cost.ChooseElevator(msg.Elevators,msg.LatestOrder.Button,msg.LatestOrder.Floor)
			e := msg.Elevators[elevId].E
			e.LightMatrix[int(msg.LatestOrder.Button)][msg.LatestOrder.Floor] = true
			e.PrintLightMatrix()
			e.ID = elevId
			fmt.Println("Hei")
			msg2 := decoding.EncodeElevatorMsg(decoding.ElevatorMsg{E:e})
			msgChanLocal <- msg2
		}
	}
}
