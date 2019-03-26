package main

import (
	//"../orders/elevio/ordStruct"
	"../MakkerModul/connector"
	"../MakkerModul/decoding"
	//"../cost"
	"fmt"
	//"os"
	//"time"
	//"math/rand"
)

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
			elevId := msg.ChooseElevator(msg.LatestOrder.Button,msg.LatestOrder.Floor)
			elevStatus := msg.Elevators[elevId]
			e := elevStatus.E
			e.LightMatrix[int(msg.LatestOrder.Button)][msg.LatestOrder.Floor] = true
			e.PrintLightMatrix()
			e.ID = elevId
			msg2 := decoding.EncodeElevatorMsg(decoding.ElevatorMsg{E:e})
			msgChanLocal <- msg2
		}
	}
}
