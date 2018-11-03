//go:generate go run github.com/UnnoTed/fileb0x assets.json

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gh0st42/goforban/forban"
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

func RunServer() {
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

func Help() {
	fmt.Println("goforban commands")
	fmt.Println("=========================\n")
	fmt.Printf(" USAGE: %v <command> [flags]\n", os.Args[0])
	fmt.Println("List of commands:")
	fmt.Println("  help        - print this help")
	fmt.Println("  serve       - start forban daemon in foreground")
	os.Exit(1)
}
func main() {
	helpCommand := flag.NewFlagSet("help", flag.ExitOnError)
	serveCommand := flag.NewFlagSet("serve", flag.ExitOnError)

	if len(os.Args) < 2 {
		Help()
	}

	switch os.Args[1] {
	case "help":
		helpCommand.Parse(os.Args[2:])
	case "serve":
		serveCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if helpCommand.Parsed() {
		Help()
	}
	if serveCommand.Parsed() {
		RunServer()
	}
}
