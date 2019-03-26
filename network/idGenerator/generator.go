package idGenerator 

import (
	"time"
	"math/rand"
)

func GetRandomID() string {
        rand.Seed(time.Now().UTC().UnixNano())
        buf := make([]byte, 8)
        for i, _ := range buf {
                buf[i] = byte(48 + rand.Intn(122-48))
        }

        return "C{" + string(buf[:]) + "}"
}

