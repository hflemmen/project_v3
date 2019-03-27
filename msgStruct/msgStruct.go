package msgStruct

import (
	"../cost"
	"../orders/elevio/ordStruct"
)

type MsgFromMaster struct {
	Id     string
	Elevators cost.ElevMap
	LightsHall ordStruct.LightType
	Number int 
}


type MsgFromElevator struct{
	ElevId string
	States ordStruct.Elevator
	Number int
}


func (msg *MsgFromMaster) UpdateLightMatrix(){
	elevator :=make(ordStruct.Elevator)
    for _,elevator_status := range msg.Elevators{
    	elevator = elevator_status.E
    	if elevator.Behaviour == ordStruct.E_DoorOpen{
    		//if door is open in a floor we can turn of the lights in the corresponding floor
    		msg.LightsHall[0][elevator.Floor] = false
    		msg.LightsHall[1][elevator.Floor] = false
    	}
    }
}