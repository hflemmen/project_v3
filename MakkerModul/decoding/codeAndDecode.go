package decoding

import (
	"../../orders/elevio/ordStruct"
	"encoding/json"
	"../../cost"
	//"fmt"
)

type ElevatorMsg struct {
	E      ordStruct.Elevator
	Number int
}

type ElevatorStatus struct {
	E              ordStruct.Elevator
	PendingUpdates bool
	CostValue      int
}

type BackupMsg struct {
	Elevators map[string]ElevatorStatus
	LatestOrder ordStruct.ButtonEvent
	Number    int
}

func DecodeElevatorMsg(str string) (outMsg ElevatorMsg) {
	json.Unmarshal([]byte(str), &outMsg)
	return
}

func EncodeElevatorMsg(msg ElevatorMsg) string {
	bytes, _ := json.Marshal(msg)
	return string(bytes)
}

func DecodeBackupMsg(str string) (outMsg BackupMsg) {
	json.Unmarshal([]byte(str), &outMsg)
	return
}

func EncodeBackupMsg(msg BackupMsg) string {
	bytes, _ := json.Marshal(msg)
	return string(bytes)
}

func (msg *BackupMsg) ChooseElevator(button ordStruct.ButtonType, floor int) string {
	minimum := 0
    i := 0 // use i to tell us if it's the first time iterating
	var elevator_id string 
	for elevator_ID,elevator_status := range msg.Elevators{
		temp :=  cost.TimeToServeRequest(elevator_status.E, button, floor)
		if i == 0 || temp < minimum {
                minimum = temp
				elevator_id = elevator_ID
		}
        i++
	}
	if (msg.Elevators != nil) {
		elev := msg.Elevators[elevator_id]
		elev.E.Order[int(button)][floor] = true
		msg.Elevators[elevator_id] = elev
	}
	return elevator_id
}