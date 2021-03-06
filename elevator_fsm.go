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

//////	ELEVATOR FSM //////
const (
	PARTNER_NAME = "msgHandler.go"
	SEND_PORT    = 55555
	RECEIVE_PORT = 44444
	NUMFLOORS    = ordStruct.NUMFLOORS
)

func main() {
	fmt.Println("Starting elevator")
	port := flag.Int("port", 15657, "number of floors")
	flag.Parse()

	if *port == 0 {
		*port = 15657
	}

	elevio.Init(fmt.Sprintf("localhost:%v", *port), NUMFLOORS)
	e := ordStruct.ElevatorInit()
	states := make(chan ordStruct.Elevator)
	//elevio.SetMotorDirection(d)

	newButton := make(chan ordStruct.ButtonEvent)
	newOrders := make(chan ordStruct.ButtonEvent)
	floorArrivals := make(chan int)
	updateLights := make(chan ordStruct.LightType)
	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel(
		PARTNER_NAME, RECEIVE_PORT, SEND_PORT)
	go elevio.PollButtons(newButton, newOrders)
	go elevio.PollFloorSensor(floorArrivals)
	go message_handler(e.Duplicate(), receiveLocal, msgChanLocal, newOrders, states, updateLights)
	elevator_fsm(e, newOrders, floorArrivals, states, updateLights, newButton)
}

func elevator_fsm(e ordStruct.Elevator, newOrders <-chan ordStruct.ButtonEvent,
	floorArrivals <-chan int, states chan ordStruct.Elevator,
	updateLights <-chan ordStruct.LightType, newButton <-chan ordStruct.ButtonEvent) {

	var doorTimer <-chan time.Time
	errorCh := make(chan string)

	prevElevator := e
	go func() {
		for err := range errorCh {
			fmt.Println("Error in FSM - ", err)
		}
	}()

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
				errorCh <- "Arrived on floor when idle!"
				fallthrough

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
				errorCh <- "Arrived on floor with doors open!"
			}

		case <-doorTimer:

			switch e.Behaviour {
			case ordStruct.E_Idle:
				//nothing

			case ordStruct.E_Moving:
				errorCh <- "Door timer when moving!"

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
		case a := <-updateLights:
			e.LightMatrix = a
			orders.UpdateLights(e)
		case <-time.After(20 * time.Millisecond):
			if prevElevator != e {
				states <- e.Duplicate()
			}
			prevElevator = e
		}
	}
}

func message_handler(elev ordStruct.Elevator, receive <-chan string, msgChan chan<- string,
	newOrders chan ordStruct.ButtonEvent, states chan ordStruct.Elevator, updateLights chan<- ordStruct.LightType) {
	msgToHandler := decoding.ElevatorMsg{E: elev, Number: 0}
	msgFromHandler := decoding.ElevatorMsg{}
	for {
		select {

		case e := <-states:
			msgToHandler.E = e
			msgToHandler.Number++
			msgChan <- decoding.EncodeElevatorMsg(msgToHandler)
		case a := <-receive:
			msg := decoding.DecodeElevatorMsg(a)
			if msg != msgFromHandler {
				buttons, floors := msgToHandler.E.Differences(msg.E)
				for i, button := range buttons {
					newOrders <- ordStruct.ButtonEvent{
						Floor:  floors[i],
						Button: ordStruct.ButtonType(button),
					}
				}
				if msg.Number > msgToHandler.Number {
					msgToHandler.Number = msg.Number
				}
				if msg.Number != -1 {
					updateLights <- msg.E.LightMatrix
				}
				msgFromHandler = msg
			}
		}
	}
}
