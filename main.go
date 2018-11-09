//go:generate go run github.com/UnnoTed/fileb0x assets.json

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gh0st42/goforban/forban"
	daemon "github.com/sevlyar/go-daemon"
	log "github.com/sirupsen/logrus"
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
	forban.DisableIPv4 = false
	forban.ListenerUDP(forban.Port)

	forban.UpdateFileIndex()

	log.Info("Starting ", forban.MyName)
	log.Info("UUIDv4: ", forban.MyUuid)

	log.Info("Index HMAC: ", forban.GetIndexHmac())

	stop := schedule(forban.Announce, 15000*time.Millisecond)

	forban.ServeHttpd()
	stop <- true
}

func Help() {
	fmt.Println("goforban commands")
	fmt.Println("=========================\n")
	fmt.Printf(" USAGE: %v <command> [flags]\n", os.Args[0])
	fmt.Println("List of commands:")
	fmt.Println("  help               - print this help")
	fmt.Println("  serve [background] - start forban daemon in foreground")
	fmt.Println("  stop               - stops forban daemon on localhost")
	fmt.Println("  share <file>       - share bundle file")
	os.Exit(1)
}
func main() {
	log.SetLevel(log.DebugLevel)

	helpCommand := flag.NewFlagSet("help", flag.ExitOnError)
	serveCommand := flag.NewFlagSet("serve", flag.ExitOnError)
	shareCommand := flag.NewFlagSet("share", flag.ExitOnError)
	stopCommand := flag.NewFlagSet("stop", flag.ExitOnError)

	if len(os.Args) < 2 {
		Help()
	}

	switch os.Args[1] {
	case "help":
		helpCommand.Parse(os.Args[2:])
	case "serve":
		serveCommand.Parse(os.Args[2:])
	case "share":
		shareCommand.Parse(os.Args[2:])
	case "stop":
		stopCommand.Parse(os.Args[2:])
	default:
		Help()
		flag.PrintDefaults()
		os.Exit(1)
	}

	if helpCommand.Parsed() {
		Help()
	}
	if stopCommand.Parsed() {
		forban.Stop()
	}
	if serveCommand.Parsed() {
		if len(os.Args) == 3 && os.Args[2] == "background" {
			cntxt := &daemon.Context{
				PidFileName: "pid",
				PidFilePerm: 0644,
				LogFileName: "log",
				LogFilePerm: 0640,
				WorkDir:     "./",
				Umask:       027,
				Args:        os.Args,
			}

			d, err := cntxt.Reborn()
			if err != nil {
				log.Fatal("Unable to run: ", err)
			}
			if d != nil {
				return
			}
			defer cntxt.Release()

		}
		RunServer()
	}
	if shareCommand.Parsed() {
		if len(os.Args) > 2 {
			forban.ShareFile(os.Args[2])
		} else {
			Help()
		}
	}
}
