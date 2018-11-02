package forban

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// ListenerUDP start listening for incoming forban announce packets
func ListenerUDP(port int) chan bool {
	stop := make(chan bool)

	go func() {
		lport := ":" + strconv.Itoa(port)
		ServerAddr, err := net.ResolveUDPAddr("udp", lport)
		CheckError(err)
		ServerConnLocal, err := net.ListenUDP("udp", ServerAddr)
		CheckError(err)
		ServerConn = ServerConnLocal
		defer ServerConn.Close()

		buf := make([]byte, 1024)
		for {
			n, addr, err := ServerConn.ReadFromUDP(buf)
			parsePkt(buf, n, addr)

			if err != nil {
				fmt.Println("Error: ", err)
			}
		}
	}()

	return stop
}

func parsePkt(pkt []byte, pktSize int, sender *net.UDPAddr) {
	recvstr := strings.TrimSpace(string(pkt[:pktSize]))
	if strings.HasPrefix(recvstr, "forban") {
		fields := strings.Split(recvstr, ";")
		if len(fields) == 7 {
			//			println("receiver announce: " + recvstr)

			var ipv4, ipv6 string
			if sender.IP.To4() != nil {
				ipv4 = sender.IP.String()
				ipv6 = ""
			} else {
				ipv4 = ""
				ipv6 = sender.IP.String()
			}

			announceNode := ForbanNode{fields[2], fields[4], fields[6], ipv4, ipv6, time.Now()}
			entry := Neighborhood[announceNode.uuid]

			if entry.node.uuid == "" {
				//println("new node")
				entry.firstSeen = time.Now()
			} else {
				//println("updated node")
			}
			if announceNode.ipv6 == "" {
				announceNode.ipv6 = entry.node.ipv6
			}
			if announceNode.ipv4 == "" {
				announceNode.ipv4 = entry.node.ipv4
			}
			entry.node = announceNode
			Neighborhood[entry.node.uuid] = entry
			if currentHmac != entry.node.hmac {
				//println("files missing")
				//fmt.Println(entry.files)
				//fmt.Println(MyFiles)
				opportunisticWorker(entry)
			}

			if len(entry.files) == 0 {
				opportunisticWorker(entry)
			}

			//println(announceNode.name, announceNode.ipv4, announceNode.ipv6)
		}
	}
}

func opportunisticWorker(entry ForbanNodeEntry) {
	//_ = "breakpoint"
	// stage 1: fetch forban/index
	addr := entry.node.ipv4

	//println(addr)
	indexurl := "http://" + addr + ":12555/s/?g=forban/index"

	resp, _ := http.Get(indexurl)
	//println(resp, err)
	body, _ := ioutil.ReadAll(resp.Body)
	forbanindex := string(body)
	defer resp.Body.Close()

	indexfiles := strings.Split(forbanindex, "\n")

	// stage 2: fetch missing files
	var filelist = []FileEntry{}

	var count int64
	count = 0
	for _, i := range indexfiles {
		if strings.HasPrefix(i, "forban/index") != true && len(i) > 0 {
			fields := strings.Split(i, ",")
			fsize, _ := strconv.ParseInt(fields[1], 10, 64)
			var curFile = FileEntry{fields[0], "", fsize}
			count += fsize
			//fmt.Println(curFile)
			//fmt.Println(fields)
			filelist = append(filelist, curFile)
			if !stringInSlice(fields[0], MyFiles) {
				fetchAndAdd(addr, fields[0])
			}
		}
	}
	entry.totalStore = count
	entry.files = filelist
	Neighborhood[entry.node.uuid] = entry
}

func fetchAndAdd(addr string, filename string) {
	b64fname := base64.StdEncoding.EncodeToString([]byte(filename))
	fileurl := "http://" + addr + ":12555/s/?g=" + b64fname + "&f=b64e"

	resp, err := http.Get(fileurl)
	if err != nil {
		fmt.Println(err)
	}
	if resp.StatusCode == 200 {
		os.MkdirAll(path.Dir(FileBasePath+"/"+filename), 0777)
		// Create the file
		out, err := os.Create(FileBasePath + "/" + filename)
		if err != nil {
			fmt.Println(err)
		}
		defer out.Close()
		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		UpdateFileIndex()
	} else {
		println(resp.StatusCode)
	}

}
