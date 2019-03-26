package connector

import (
	//"bytes"
	//"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

/*
	TODO
	[ ] Errorhandling
	[ ] Sikkerhetsnett for startProgram
	[ ] En message struct som er mer brukbar enn en string
	[ ] Rydding av variabelnavn og struktur
	[ ] Kommentarer

*/
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func listenLocal(port int, rx chan string, aliveMsgChan chan bool) {
	localAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	checkErr(err)
	connection, err := net.ListenUDP("udp", localAddress)
	checkErr(err)
	buffer := make([]byte, 1024)
	var length int
	var msg string
	prevMsg := ""

	for err == nil {
		defer connection.Close()
		length, _, err = connection.ReadFromUDP(buffer)
		msg = string(buffer[:length])
		msg = strings.Replace(msg, "\x00", "", -1)
		switch msg {
		case "":
		case "I'm alive!":
			aliveMsgChan <- true
		default:
			if prevMsg != msg {
				rx <- msg
				prevMsg = msg
			}
		}
	}
	fmt.Println("Error in local listener, ", err)
}

func sendLocal(port int, tx chan string) {
	dst, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	checkErr(err)
	conn, err := net.DialUDP("udp", nil, dst)
	checkErr(err)
	for msg := range tx {
		buffer := make([]byte, 1024)
		copy(buffer[:], msg)
		_, err = conn.Write(buffer)
		if err != nil {
			//	break
		}
	}
	fmt.Println("Error in local sender, ", err)
}

func EstablishLocalTunnel(partnerName string, rxPort int, txPort int) (chan string, chan string) {
	msgChan := make(chan string)

	go sendLocal(txPort, msgChan)
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			msgChan <- "I'm alive!"
		}
	}()

	receive := make(chan string)
	partnerAliveCh := make(chan bool)

	go listenLocal(rxPort, receive, partnerAliveCh)
	go keepAlive(partnerName, receive, partnerAliveCh)
	return receive, msgChan
}

func keepAlive(partnerName string, rx chan string, parnerAliveCh chan bool) {
	timeOutCnt := -2000
	hasConnection := false
	incrementWatchDogCnt := make(chan bool)
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			incrementWatchDogCnt <- true
		}
	}()
	go func() {
		for {
			select {
			case <-incrementWatchDogCnt:
				switch {
				case timeOutCnt > 1000:
					if hasConnection {
						fmt.Printf("Connection lost with %v\n", partnerName)
						rx <- "Connection lost"
					}
					hasConnection = false
					go startProgram(partnerName)
					timeOutCnt = -30000
				case timeOutCnt > 200:
					fmt.Printf("Watchdog for %v, at %v ms\n", partnerName, timeOutCnt)
					fallthrough
				default:
					timeOutCnt += 200
				}
			case <-parnerAliveCh:
				timeOutCnt = 0
				if !hasConnection {
					fmt.Printf("Connection established with %v\n", partnerName)
					rx <- "Connection established"
					hasConnection = true
				}
			}
		}
	}()
}

func startProgram(name string) {
	if !strings.HasSuffix(name, ".go") {
		name = name + ".go"
	}

	cmd := exec.Command("gnome-terminal", "-x", "go", "run", name)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", "start", "go", "run", name)
		cmd.Stdout = os.Stdout //trengs disse?
		cmd.Stderr = os.Stderr //trengs disse?
	}
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
