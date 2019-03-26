package cost

import (
	"../orders"
	//."../orders/elevio"
	"../orders/elevio/ordStruct"
)

//constants for doing the calculations in TimeToServeRequest
//not to be used anywhere else
const TRAVEL_TIME = 2
const DOOR_OPEN_TIME = 3

type ElevatorStatus struct {
    E              ordStruct.Elevator
    PendingUpdates bool
    CostValue      int
}
type ElevMap struct {
    Elevators map[string]ElevatorStatus
}


func TimeToServeRequest(e_old ordStruct.Elevator, button ordStruct.ButtonType, floor int) int {
    e := e_old
    e.Order[button][floor] = true
    duration := 0
    
    switch e.Behaviour{
    case ordStruct.E_Idle:
        e.Dir = orders.ChooseDirection(e)
        if(e.Dir == ordStruct.MD_Stop){
            return duration
        }
        break
    case ordStruct.E_Moving:
        duration += TRAVEL_TIME/2;
        e.Floor += int(e.Dir)
        break
    case ordStruct.E_DoorOpen:
        duration -= DOOR_OPEN_TIME/2
    }

    for {
        if(orders.ShouldStop(e) == true){
            e = ClearOrdersAtCurrentFloorCost(e)
            if(e.Floor == floor){
                return duration
            }
            duration += DOOR_OPEN_TIME
            e.Dir = orders.ChooseDirection(e)
        }
        e.Floor += int(e.Dir)
        duration += TRAVEL_TIME
    }
}


//simulated clear order function to be sure to not mess with actual elevator states
func ClearOrdersAtCurrentFloorCost(e ordStruct.Elevator) ordStruct.Elevator {
	e2 := e.Duplicate()
	for btn := 0; btn < 3; btn++ {
		if e2.Order[btn][e2.Floor] != false {
			e2.Order[btn][e2.Floor] = false
		}
	}
	return e2;
}
func (elevMap *ElevMap) ChooseElevator(button ordStruct.ButtonType, floor int) string {
    minimum := 0
    i := 0 // use i to tell us if it's the first time iterating
    var elevator_id string 
    for elevator_ID,elevator_status := range elevMap.Elevators{
        temp :=  TimeToServeRequest(elevator_status.E, button, floor)
        if i == 0 || temp < minimum {
                minimum = temp
                elevator_id = elevator_ID
        }
        i++
    }
    /*
    if (elevMap.Elevators != nil) {
        elev := elevMap.Elevators[elevator_id]
        elev.E.Order[int(button)][floor] = true
        elevMap.Elevators[elevator_id] = elev
    }
    */
    return elevator_id
}


