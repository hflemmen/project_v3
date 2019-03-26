package main

import "fmt"
import "./orders/elevio"
import	"./orders/elevio/ordStruct"
import "./orders"
import "time"

func main() {
	elevio.Init("localhost:15657", 4)
	e := ordStruct.ElevatorInit(1,4)
	e.Floor = 0
	e.Dir = 0
	LightMatrixZero := ordStruct.LightType{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}
	LightMatrixOne := ordStruct.LightType{
		{1, 1, 1, 1},
		{1, 1, 1, 1},
		{1, 1, 1, 1},
	}
	fmt.Println("hello")
	for {
		e.LightMatrix = LightMatrixZero
		orders.UpdateLights(e)
		time.Sleep(2*time.Second)
		e.LightMatrix = LightMatrixOne
		orders.UpdateLights(e)
		time.Sleep(2*time.Second)
	}
}
	
