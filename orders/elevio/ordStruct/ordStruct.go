package ordStruct

import "fmt"
import "time"

const DOOR_OPEN_TIME = 3 * time.Second
const NUMFLOORS = 4

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

type LightType [2][NUMFLOORS]bool

type Elevator struct {
	Floor       int
	Dir         MotorDirection
	Order       [3][NUMFLOORS]bool
	Behaviour   behaviourType
	LightMatrix LightType
}

func ElevatorInit() (e Elevator) {
	e = Elevator{Dir: 0}
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
		for j := 0; j < NUMFLOORS; j++ {
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
func (e *Elevator) CheckOrderUpdate() ButtonEvent {
	for btn := 0; btn < 3; btn++ {
		for floor := 0; floor < NUMFLOORS; floor++ {
			if e.Order[btn][floor] == false && e.LightMatrix[btn][floor] == true {
				return ButtonEvent{Button: ButtonType(btn), Floor: floor}
			}
		}
	}
	return ButtonEvent{Floor: -1}
}
