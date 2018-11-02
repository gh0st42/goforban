//go:generate go run github.com/UnnoTed/fileb0x assets.json

package main

import (
	"fmt"
	"time"

	"goforban/forban"
)

func schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func main() {
	forban.InitForban()
	//forban.Interfaces = []string{"en1"}
	//forban.Interfaces = []string{"wlan0"}
	forban.Interfaces = []string{"eth0"}
	forban.DisableIPv6 = true
	forban.ListenerUDP(forban.Port)

	forban.UpdateFileIndex()

	println("Starting ", forban.MyName)
	fmt.Printf("UUIDv4: %s\n", forban.MyUuid)

	println(forban.GetIndexHmac())

	stop := schedule(forban.Announce, 5000*time.Millisecond)

	forban.ServeHttpd()
	stop <- true
}
