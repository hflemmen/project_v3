package ordStruct

import "fmt"
import "time"

const DOOR_OPEN_TIME = 3 * time.Second

type behaviourType int

const (
	E_Moving   behaviourType = 2
	E_Idle                   = 1
	E_DoorOpen               = 0
)

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type OrderType int

const (
	OT_Up   OrderType = 1
	OT_Down           = -1
	OT_Cab            = 0
)

type ButtonEvent struct {
	Floor     int
	Direction int
	Button    ButtonType
}

type LightType [2][4]bool

type Elevator struct {
	ID          string
	Floor       int
	NumFloors   int
	Dir         MotorDirection
	Order       [3][4]bool
	Behaviour   behaviourType
	LightMatrix LightType
}

func ElevatorInit(id string, floors int) (e Elevator) {
	e = Elevator{ID: id, NumFloors: floors, Dir: 0}
	return
}
func (e *Elevator) Duplicate() (e2 Elevator) {
	e2 = *e
	return
}

func (e *Elevator) AddOrder(button int, floor int) {
	e.Order[button][floor] = true
}

func (e *Elevator) RemoveOrder(button ButtonType, floor int) {
	e.Order[int(button)][floor] = false
}

func (e *Elevator) Differences(e2 Elevator) ([]int, []int) {
	buttons := make([]int, 0)
	floors := make([]int, 0)
	for i := 0; i < 3; i++ {
		for j := 0; j < e.NumFloors; j++ {
			if e2.Order[i][j] == true && e.Order[i][j] == false {
				buttons = append(buttons, i)
				floors = append(floors, j)
			}
		}
	}
	return buttons, floors
}

func (e *Elevator) ClearOrders() {
	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			e.Order[i][j] = false
		}
	}
}

func (e *Elevator) PrintOrders() {
	fmt.Print("[ ")
	for i := 0; i < 3; i++ {
		fmt.Println()
		for j := 0; j < 4; j++ {
			fmt.Print(e.Order[i][j], " ")
		}
	}
	fmt.Print("]\n")
}
func (e *Elevator) PrintLightMatrix() {
	fmt.Print("[ ")
	for i := 0; i < 2; i++ {
		fmt.Println()
		for j := 0; j < 4; j++ {
			fmt.Print(e.LightMatrix[i][j], " ")
		}
	}
	fmt.Print("]\n")
}

/*
func (e *Elevator) LightUpdate(button ButtonType,floor int, on bool, receiverLights chan<- LightEvent){
	if button == BT_Cab {
		fmt.Println("Error: external light can not be CAB!")
	}
	receiverLights<-LightEvent{Floor: floor, On: on, Button: button}

	for i := 0; i < 2; i++ {
		for j := 0; j < 4; j++ {
				e.Lights[i][j]|= externalList[i][j]
			}
		}
	}

}
*/
