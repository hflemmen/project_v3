package decoding

import (
	"../../orders/elevio/ordStruct"
	"encoding/json"
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
	number    int
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
