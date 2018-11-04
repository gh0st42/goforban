//go:generate go run github.com/UnnoTed/fileb0x assets.json

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

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
	forban.DisableIPv4 = false
	forban.ListenerUDP(forban.Port)

	forban.UpdateFileIndex()

	log.Info("Starting ", forban.MyName)
	log.Info("UUIDv4: ", forban.MyUuid)

	log.Info("Index HMAC: ", forban.GetIndexHmac())

	stop := schedule(forban.Announce, 5000*time.Millisecond)

	forban.ServeHttpd()
	stop <- true
}

func ShareFile(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("File does not exist")
	} else {
		fmt.Println("File exists")
	}
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	res, err := http.Post("http://127.0.0.1:12555/upload", "binary/octet-stream", file)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	message, _ := ioutil.ReadAll(res.Body)
	println(string(message))

	println("DONE")
}
func Help() {
	fmt.Println("goforban commands")
	fmt.Println("=========================\n")
	fmt.Printf(" USAGE: %v <command> [flags]\n", os.Args[0])
	fmt.Println("List of commands:")
	fmt.Println("  help         - print this help")
	fmt.Println("  serve        - start forban daemon in foreground")
	fmt.Println("  share <file> - share bundle file")
	os.Exit(1)
}
func main() {
	log.SetLevel(log.DebugLevel)

	helpCommand := flag.NewFlagSet("help", flag.ExitOnError)
	serveCommand := flag.NewFlagSet("serve", flag.ExitOnError)
	shareCommand := flag.NewFlagSet("share", flag.ExitOnError)

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
	default:
		Help()
		flag.PrintDefaults()
		os.Exit(1)
	}

	if helpCommand.Parsed() {
		Help()
	}
	if serveCommand.Parsed() {
		RunServer()
	}
	if shareCommand.Parsed() {
		if len(os.Args) > 2 {
			ShareFile(os.Args[2])
		} else {
			Help()
		}
	}
}
