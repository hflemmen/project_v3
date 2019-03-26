package orders

import . "./elevio/ordStruct"
import "./elevio"

//import "time"

func orderAbove(e Elevator) bool {
	for floor := e.Floor + 1; floor < e.NumFloors; floor++ {
		for btn := 0; btn < 3; btn++ {
			if e.Order[btn][floor] {
				return true
			}
		}
	}
	return false
}
func orderBelow(e Elevator) bool {
	for floor := 0; floor < e.Floor; floor++ {
		for btn := 0; btn < 3; btn++ {
			if e.Order[btn][floor] {
				return true
			}
		}
	}
	return false
}
func ChooseDirection(e Elevator) MotorDirection {
	switch e.Dir {
	case MD_Down:
		fallthrough
	case MD_Up:
		if orderAbove(e) {
			return MD_Up
		} else if orderBelow(e) {
			return MD_Down
		} else {
			return MD_Stop
		}
	case MD_Stop:
		if orderBelow(e) {
			return MD_Down
		} else if orderAbove(e) {
			return MD_Up
		} else {
			return MD_Stop
		}
	default:
		return MD_Stop
	}

}

func ShouldStop(e Elevator) bool {
	switch e.Dir {
	case MD_Down:
		return e.Order[BT_HallDown][e.Floor] ||
			e.Order[BT_Cab][e.Floor] ||
			!orderBelow(e)
	case MD_Up:
		return e.Order[BT_HallUp][e.Floor] ||
			e.Order[BT_Cab][e.Floor] ||
			!orderAbove(e)
	case MD_Stop:
		fallthrough
	default:
		return true
	}
}

func ClearOrdersAtCurrentFloor(e Elevator) Elevator {
	e2 := e.Duplicate()
	for btn := 0; btn < 3; btn++ {
		if e2.Order[btn][e2.Floor] {
			e2.Order[btn][e2.Floor] = false
		}
	}
	return e2
}
func ClearLightsAtCurrentFloor(e Elevator) Elevator {
	e2 := e.Duplicate()
	for btn := 0; btn < 2; btn++ {
		if e2.LightMatrix[btn][e2.Floor] {
			e2.LightMatrix[btn][e2.Floor] = false
		}
	}
	elevio.SetButtonLamp(BT_Cab, e.Floor, false)
	return e2
}

func UpdateLights(e Elevator) {
	for floor := 0; floor < e.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			if e.LightMatrix[btn][floor] {
				elevio.SetButtonLamp(ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(ButtonType(btn), floor, false)
			}
		}
		if e.Order[int(BT_Cab)][floor] {
			elevio.SetButtonLamp(BT_Cab, floor, true)
		} else {
			elevio.SetButtonLamp(BT_Cab, floor, false)
		}
	}
}
