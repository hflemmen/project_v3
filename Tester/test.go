package main

import (
	. "./cost"
	. "./orders/ordStruct"
	. "./orders/ordStruct/elevio"
	"fmt"
	"time"
)

func main() {
	elevator_1 := Elevator{
		ID:        1,
		Floor:     0,
		NumFloors: 4,
		Dir:       MD_Up,
		Order: [3][4]int{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		OpenDoorTime: 3 * time.Second,
		Behaviour:    E_Idle,
	}

	elevator_2 := Elevator{
		ID:        2,
		Floor:     1,
		NumFloors: 4,
		Dir:       MD_Down,
		Order: [3][4]int{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{1, 0, 0, 0},
		},
		OpenDoorTime: 3 * time.Second,
		Behaviour:    E_Moving,
	}
	var elevators []Elevator
	elevators = append(elevators, elevator_1)
	elevators = append(elevators, elevator_1)

	time_Elevator_1 := TimeToServeRequest(elevator_1, 1, 3)
	time_Elevator_2 := TimeToServeRequest(elevator_2, 1, 3)
	identification := ChooseElevator(elevators, 1, 3)
	fmt.Println("Elevator 1 uses: ", time_Elevator_1)
	fmt.Println("Elevator 2 uses: ", time_Elevator_2)
	fmt.Println("We choose elevator: ", identification, " for the order")
}
