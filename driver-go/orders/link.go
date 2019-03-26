package orders

import	"fmt"

type OrderVal struct {
	floor int
	dir int
}

type Node struct {
	OrderVal
	next *Node
}

type OrderList struct {
	head, tail *Node
}

func (l *OrderList) AddOrder(floor int, dir int){
	//fmt.Printf("Adding at %v\n", floor)
	n := &Node{OrderVal{floor, dir}, nil}
	if l.head == nil{
		l.head, l.tail = n, n
		return
	}
	if l.head.floor > floor{
		n.next = l.head
		l.head = n
		return
	}
	pvPtr := l.head
	for ptr := l.head; ptr != nil; ptr = ptr.next{
		if ptr.floor == floor {
			if ptr.dir != dir {
				ptr.dir = 0
			}
			return
		}
		if ptr.floor > floor {
			pvPtr.next = n
			n.next = ptr
			return
		}
		pvPtr = ptr
	}
	l.tail.next = n
	l.tail = n
}

func (l *OrderList) RemoveOrder(floor int) {
	//l.PrintTechnical()
	//fmt.Printf("Removing %v\n", floor)
	if l.head == nil{
		return
	}
	if l.head.floor == floor{
		if l.head == l.tail{
			l.head = nil
			l.tail = nil
			return
		}
		l.head = l.head.next
		return
	}
	if l.head == l.tail {
		return
	}
	ptr := l.head
	for ptr.floor < floor && ptr.next != l.tail{
		if ptr.next.floor == floor {
			ptr.next = ptr.next.next
			return
		}
		ptr = ptr.next
	}
	if l.tail.floor == floor {
		l.tail = ptr
		ptr.next = nil
	}

}

func (l *OrderList) RemoveAll(){
	l.head, l.tail = nil, nil
}

func (l *OrderList) FindDir(currFloor int, dir int) int{
	if l.head == nil {
		return 0
	}
	if dir == 0 {
		diff := (currFloor - l.head.floor)
		if diff > 0{
			return -1
		}else if (diff < 0){
			return 1
		}else{
			return 0
		}
	}
	if dir == -1 && currFloor <= l.head.floor{
		return 1
	}
	if dir == 1 && currFloor >= l.tail.floor{
		return -1
	}
	return dir
}

func (l *OrderList) CheckIfInList(currFloor int, dir int) bool{
	if l.head == nil || l.tail == nil{
		return false
	}
	if dir == -1 && currFloor <= l.head.floor{
		return true
	}
	if dir == 1 && currFloor >= l.tail.floor{
		return true
	}
	for ptr := l.head; ptr != nil && ptr.floor <= currFloor; ptr = ptr.next{
		if ptr.floor == currFloor{
			if ptr.dir == 0 || ptr.dir == dir{
				return true
			}
			return false
		}
	}
	return false
}

func (l1 *OrderList) IntegrateList(l2 *OrderList) {
	for ptr := l2.head; ptr != nil; ptr = ptr.next{
		l1.AddOrder(ptr.floor, ptr.dir)
	}
}


func (l *OrderList) PrintList(){
	if (l.head != nil && l.tail != nil){
		fmt.Printf("Head: F{%v}D{%v}\nTail: F{%v}D{%v}\n",
		l.head.floor, l.head.dir, l.tail.floor, l.tail.dir)
		fmt.Printf("List: ")
		for ptr := l.head; ptr != nil; ptr = ptr.next{
			fmt.Printf("F{%v}D{%v}->", ptr.floor, ptr.dir)
		}
		fmt.Printf("END\n")
	}else{
		fmt.Printf("List is empty\n")
	}
}

func (l *OrderList) PrintTechnical(){
	fmt.Printf("Head: F{%v}{%p}\tTail: F{%v}{%p}\n", l.head.floor, l.head, l.tail.floor, l.tail)
	fmt.Printf("List: ")
	for ptr := l.head; ptr != nil; ptr = ptr.next{
		fmt.Printf("F{%v}{%p}->", ptr.floor, ptr)
	}
	fmt.Printf("END\n\n")
}

func (l *OrderList) StressTest(n int){
	for i := 0; i < n; i++{
		l.AddOrder(i, 0)
	}
	for i := 0; i < n; i++{
		l.RemoveOrder(i)
	}

	for i := n; i >= 0; i--{
		l.AddOrder(i, 0)
	}
	for i := n; i >= 0; i--{
		l.RemoveOrder(i)
	}

	//for i := 0; i < n; i++{
	//	l.AddOrder(rand.Intn(n), (1-rand.Intn(3)))
	//	l.RemoveOrder(rand.Intn(l.tail.floor))
	//}
	l.PrintList()
	l.RemoveAll()
}
