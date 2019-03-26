package cost

import (

	."../orders"
	."../orders/elevio"
	."../orders/elevio/ordStruct"
)

const TRAVEL_TIME = 2
const DOOR_OPEN_TIME = 3


func TimeToServeRequest(e_old Elevator, button ButtonType, floor int) int {
    e := e_old
    e.Order[button][floor] = 1
    duration := 0
    
    switch e.Behaviour{
    case E_Idle:
        e.Dir = ChooseDirection(e)
        if(e.Dir == MD_Stop){
            return duration
        }
        break
    case E_Moving:
        duration += TRAVEL_TIME/2;
        e.Floor += int(e.Dir)
        break
    case E_DoorOpen:
        duration -= DOOR_OPEN_TIME/2
    }

    for {
        if(ShouldStop(e) == true){
            e = ClearOrdersAtCurrentFloor_Cost(e)
            if(e.Floor == floor){
                return duration
            }
            duration += DOOR_OPEN_TIME
            e.Dir = ChooseDirection(e)
        }
        e.Floor += int(e.Dir)
        duration += TRAVEL_TIME
    }
}


func ClearOrdersAtCurrentFloor_Cost(e Elevator) Elevator {
	e2 := e.Duplicate()
	for btn := 0; btn < 3; btn++ {
		if e2.Order[btn][e2.Floor] != 0 {
			e2.Order[btn][e2.Floor] = 0
		}
	}
	return e2;
}


func ChooseElevator(elevators []Elevator, button ButtonType, floor int) int {
	minimum := 0
	var elevator_id int 
	for i,elevator := range elevators{
		temp :=  TimeToServeRequest(elevator, button, floor)
		if i == 0 || temp < minimum {
                minimum = temp
				elevator_id = elevator.ID 
		}
	}
	return elevator_id
}