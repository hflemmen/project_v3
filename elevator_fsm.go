package main

import (
	"./MakkerModul/connector"
	"./MakkerModul/decoding"
	"./orders"
	"./orders/elevio"
	"./orders/elevio/ordStruct"
	"flag"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting elevator")
	floors := flag.Int("floors", 4, "number of floors")
	port := flag.Int("port", 15657, "number of floors")
	flag.Parse()
	numFloors := *floors

	if numFloors == 0 {
		numFloors = 4
	}
	if *port == 0 {
		*port = 15657
	}

	elevio.Init(fmt.Sprintf("localhost:%v", *port), numFloors)
	e := ordStruct.ElevatorInit("elev", numFloors)
	states := make(chan ordStruct.Elevator)
	//elevio.SetMotorDirection(d)

	newButton := make(chan ordStruct.ButtonEvent)
	newOrders := make(chan ordStruct.ButtonEvent)
	floorArrivals := make(chan int)
	updateLights := make(chan ordStruct.LightType)
	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel(
		"msgHandler.go", 44444, 55555)
	go elevio.PollButtons(newButton, newOrders)
	go elevio.PollFloorSensor(floorArrivals)
	go message_handler(receiveLocal, msgChanLocal, newOrders, states, updateLights)
	elevator_fsm(e, newOrders, floorArrivals, states, updateLights, newButton)
}

func elevator_fsm(e ordStruct.Elevator, newOrders <-chan ordStruct.ButtonEvent,
	floorArrivals <-chan int, states chan ordStruct.Elevator,
	updateLights <-chan ordStruct.LightType, newButton <-chan ordStruct.ButtonEvent) {

	var doorTimer <-chan time.Time
	f := elevio.GetFloor()
	if f == -1 {
		elevio.SetMotorDirection(ordStruct.MD_Down)
		e.Behaviour = ordStruct.E_Moving
	} else {
		elevio.SetMotorDirection(ordStruct.MD_Stop)
		doorTimer = time.After(ordStruct.DOOR_OPEN_TIME)
		e.Floor = f
		e.Behaviour = ordStruct.E_DoorOpen
	}
	for {
		prevElevator := e
	sel:
		select {
		case a := <-newButton:
			switch e.Behaviour {
			case ordStruct.E_Idle:
				if e.Floor == a.Floor {
					elevio.SetDoorOpenLamp(true)
					doorTimer = time.After(ordStruct.DOOR_OPEN_TIME)
					e.Behaviour = ordStruct.E_DoorOpen
				} else {
					if a.Button == ordStruct.BT_Cab {
						e.Order[a.Button][a.Floor] = true
					} else {
						e.LightMatrix[a.Button][a.Floor] = true
					}
				}

			case ordStruct.E_Moving:
				if a.Button == ordStruct.BT_Cab {
					e.Order[a.Button][a.Floor] = true
				} else {
					e.LightMatrix[a.Button][a.Floor] = true
				}

			case ordStruct.E_DoorOpen:
				if e.Floor == a.Floor {
					doorTimer = time.After(ordStruct.DOOR_OPEN_TIME)
				} else {
					if a.Button == ordStruct.BT_Cab {
						e.Order[a.Button][a.Floor] = true
					} else {
						e.LightMatrix[a.Button][a.Floor] = true
					}
				}
			}
			if e != prevElevator {
				states <- e.Duplicate()
			}
		case a := <-newOrders:
			switch e.Behaviour {
			case ordStruct.E_Idle:
				if e.Floor == a.Floor {
					elevio.SetDoorOpenLamp(true)
					doorTimer = time.After(ordStruct.DOOR_OPEN_TIME)
					e.Behaviour = ordStruct.E_DoorOpen
				} else {
					e.Order[a.Button][a.Floor] = true
					if a.Button == ordStruct.BT_Cab {
						elevio.SetButtonLamp(ordStruct.BT_Cab, a.Floor, true)
					}
					e.Dir = orders.ChooseDirection(e)
					elevio.SetMotorDirection(e.Dir)
					e.Behaviour = ordStruct.E_Moving
				}

			case ordStruct.E_Moving:
				e.Order[a.Button][a.Floor] = true
				if a.Button == ordStruct.BT_Cab {
					elevio.SetButtonLamp(ordStruct.BT_Cab, a.Floor, true)
					break sel
				}

			case ordStruct.E_DoorOpen:
				if e.Floor == a.Floor {
					doorTimer = time.After(ordStruct.DOOR_OPEN_TIME)
				} else {
					e.Order[a.Button][a.Floor] = true
					if a.Button == ordStruct.BT_Cab {
						elevio.SetButtonLamp(ordStruct.BT_Cab, a.Floor, true)
					}
				}
			}
		case a := <-floorArrivals:
			e.Floor = a
			elevio.SetFloorIndicator(e.Floor)

			switch e.Behaviour {
			case ordStruct.E_Idle:
				//do nothing
			case ordStruct.E_Moving:
				if orders.ShouldStop(e) {
					elevio.SetMotorDirection(ordStruct.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					e = orders.ClearOrdersAtCurrentFloor(e)
					e = orders.ClearLightsAtCurrentFloor(e)
					doorTimer = time.After(ordStruct.DOOR_OPEN_TIME)
					e.Behaviour = ordStruct.E_DoorOpen
				}
			case ordStruct.E_DoorOpen:
				//nothing
			}
			if e != prevElevator {
				states <- e.Duplicate()
			}
		case <-doorTimer:

			switch e.Behaviour {
			case ordStruct.E_Idle:
				fallthrough
			case ordStruct.E_Moving:
				fallthrough
			case ordStruct.E_DoorOpen:
				e.Dir = orders.ChooseDirection(e)
				elevio.SetDoorOpenLamp(false)
				elevio.SetMotorDirection(e.Dir)
				if e.Dir == ordStruct.MD_Stop {
					e.Behaviour = ordStruct.E_Idle
				} else {
					e.Behaviour = ordStruct.E_Moving
				}
			}
			if prevElevator != e {
				states <- e.Duplicate()
			}
		case a := <-updateLights:
			e.LightMatrix = a
			orders.UpdateLights(e)
		}
	}
}

func message_handler(receive <-chan string, msgChan chan<- string,
	newOrders chan ordStruct.ButtonEvent, states chan ordStruct.Elevator, updateLights chan<- ordStruct.LightType) {
	msgToHandler := decoding.ElevatorMsg{Number: 1}
	for {
		select {
		case a := <-receive:
			msg := decoding.DecodeElevatorMsg(a)
			buttons, floors := msgToHandler.E.Differences(msg.E)
			for i, _ := range buttons {
				newOrders <- ordStruct.ButtonEvent{
					Floor:  floors[i],
					Button: ordStruct.ButtonType(buttons[i]),
				}
			}
			if msg.Number > msgToHandler.Number {
				msgToHandler.Number = msg.Number
			}
			updateLights <- msg.E.LightMatrix
		case e := <-states:
			msgToHandler.E = e
			msgToHandler.Number++
			msgChan <- decoding.EncodeElevatorMsg(msgToHandler)
		}
	}
}
