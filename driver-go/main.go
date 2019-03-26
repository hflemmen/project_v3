package main

import "./elevio"
import "./orders"
import "fmt"

//Er kun utenfor main for printfunksjonen

var currFloor int
var d elevio.MotorDirection

//Er også utenfor for 
var list orders.OrderList
var moving bool = false

func main(){

    numFloors := 4
    elevio.Init("localhost:44444", numFloors)

    for i := 0; i < numFloors; i++{
	    elevio.SetButtonLampOnFloor(i, false)
    }

    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)

    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)


    for {
        select {
        case a := <-drv_buttons:
            fmt.Printf("Button pressed %+v\n", a)
            elevio.SetButtonLamp(a.Button, a.Floor, true)
	    list.AddOrder(a.Floor, a.Direction)
	    if !moving && a.Floor == currFloor{
		list.RemoveOrder(a.Floor)
		elevio.SetButtonLampOnFloor(a.Floor, false)
	    }else{
		updateAndSetMotorDirection(currFloor)
	    }
        case a := <-drv_floors:
	    currFloor = a
	    if list.CheckIfInList(a, int(d)) || int(d) == 0{
	        list.RemoveOrder(a)
		elevio.SetButtonLampOnFloor(a, false)
	    }
	    updateAndSetMotorDirection(currFloor)
        case a := <-drv_obstr:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }

        case a := <-drv_stop:
            fmt.Printf("%+v\n", a)
	    list.RemoveAll()
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
		}
            }
	}

    }
}

func printStatus(){
	fmt.Printf("\n\n######################\n")
	fmt.Printf("########STATUS########\n\n")
	fmt.Printf("CurrFloor: %v\n", currFloor)
	fmt.Printf("Direction: %v\n", int(d))
	list.PrintList()
	fmt.Printf("#########STOP#########\n")
	fmt.Printf("######################\n")
}

func updateAndSetMotorDirection(floor int) {
	d = elevio.MotorDirection(list.FindDir(floor,int(d)))
	printStatus()
	if int(d) == 0 {
		moving = false
	}else{
		moving = true
	}
	elevio.SetMotorDirection(d)
}
