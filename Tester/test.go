package main

import (
	"../cost"
	"../orders/elevio/ordStruct"
	"../MakkerModul/decoding"
	"fmt"
)

func main() {
	elevator_1 := ordStruct.Elevator{
		ID:        "1",
		Floor:     0,
		NumFloors: 4,
		Dir:       ordStruct.MD_Up,
		Order: [3][4]int{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		Behaviour:    ordStruct.E_Idle,
	}

	elevator_status_1 := decoding.ElevatorStatus{E: elevator_1}

	elevator_2 := ordStruct.Elevator{
		ID:        "2",
		Floor:     1,
		NumFloors: 4,
		Dir:       ordStruct.MD_Down,
		Order: [3][4]int{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{1, 0, 0, 0},
		},
		Behaviour:    ordStruct.E_Moving,
	}

	elevator_status_2 := decoding.ElevatorStatus{E: elevator_2}

	elevators:= make(map[string]decoding.ElevatorStatus)
	elevators["1"] = elevator_status_1
	elevators["2"] = elevator_status_2

	time_Elevator_1 := cost.TimeToServeRequest(elevator_status_1.E, 1, 3)
	time_Elevator_2 := cost.TimeToServeRequest(elevator_status_2.E, 1, 3)
	identification := cost.ChooseElevator(elevators, 1, 3)
	fmt.Println("Elevator 1 uses: ", time_Elevator_1)
	fmt.Println("Elevator 2 uses: ", time_Elevator_2)
	fmt.Println("We choose elevator: ", identification, " for the order")
}
