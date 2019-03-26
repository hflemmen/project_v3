package main

import (
	"./MakkerModul/connector"
	"./MakkerModul/decoding"
	"./orders"
	"./orders/elevio"
	"./orders/elevio/ordStruct"
	"fmt"
	"time"
)

func main() {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)
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
	go connectionMsgHandler(receiveLocal, msgChanLocal, newOrders, states, updateLights)
	elevator_fsm(e, newOrders, floorArrivals, states, updateLights, newButton)
}

func elevator_fsm(e ordStruct.Elevator, newOrders <-chan ordStruct.ButtonEvent,
	floorArrivals <-chan int, states chan ordStruct.Elevator, updateLights <-chan ordStruct.LightType,
	newButton <-chan ordStruct.ButtonEvent) {
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
	states <- e.Duplicate()
	for {
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
			states <- e.Duplicate()
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
			states <- e
		case <-doorTimer:
			switch e.Behaviour {
			case ordStruct.E_Idle:
				//nothing
			case ordStruct.E_Moving:
				//nothing
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
			states <- e
		case a := <-updateLights:
			e.LightMatrix = a
			orders.UpdateLights(e)

		}
	}
}

func connectionMsgHandler(receive <-chan string, msgChan chan<- string,
	newOrders chan ordStruct.ButtonEvent, states chan ordStruct.Elevator, updateLights chan<- ordStruct.LightType) {
	msgToHandler := decoding.ElevatorMsg{Number: 1}
	//hasConnection := false
	for {
		select {
		case a := <-receive:
			msg := decoding.DecodeElevatorMsg(a)
			fmt.Println(msg.E.ID)
			buttons, floors := msgToHandler.E.Differences(msg.E)
			for i := 0; i < len(buttons); i++ {
				newOrders <- ordStruct.ButtonEvent{Floor: floors[i],
					Button: ordStruct.ButtonType(buttons[i])}
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
