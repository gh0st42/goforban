package forban

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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
			if err != nil {
				log.Error("NET Error: ", err)
			}

			log.Debug("NET Received announcement from ", addr)
			parsePkt(buf, n, addr)
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
			newNode := false

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
				log.Info("NET New node discovered: ", announceNode)
				entry.firstSeen = time.Now()
				newNode = true
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
				count, ok := HmacIgnoreList[entry.node.hmac]
				if ok {
					log.Debug("NET Known HMAC mismatch, ignoring")
					HmacIgnoreList[entry.node.hmac]--
					if count == 1 {
						delete(HmacIgnoreList, entry.node.hmac)
					}
				} else {
					HmacIgnoreList[entry.node.hmac] = IgnoreCount
					log.Debug("NET HMAC mismatch: ", currentHmac, " ", entry.node.hmac)
					opportunisticWorker(entry)
				}
			}

			if len(entry.files) == 0 && newNode {
				log.Debug("NET Unknown file count for node ", entry.node.name, ", requesting index")
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

	resp, err := http.Get(indexurl)
	if err != nil {
		log.Error("Error fetching index from ", addr, " : ", err)

		return
	}
	if resp.StatusCode == 200 {
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
					_, ok := DownloadQueue[addr+fields[0]]
					if ok == false {
						go fetchAndAdd(addr, fields[0])
						DownloadQueue[addr+fields[0]] = true
					} else {
						log.Debug("NET Download still in progress")
					}
				}
			}
		}
		entry.totalStore = count
		entry.files = filelist
		Neighborhood[entry.node.uuid] = entry
	} else {
		log.Error("NET ", resp.StatusCode, " ", indexurl)
	}
}

func fetchAndAdd(addr string, filename string) {
	b64fname := base64.StdEncoding.EncodeToString([]byte(filename))
	fileurl := "http://" + addr + ":12555/s/?g=" + b64fname + "&f=b64e"

	log.Debug("NET Fetching ", fileurl)

	resp, err := http.Get(fileurl)
	if err != nil {
		log.Error(err)
		delete(DownloadQueue, addr+filename)
		return
	}
	defer resp.Body.Close()
	log.Debug("NET GET ", resp)
	if resp.StatusCode == http.StatusOK {
		os.MkdirAll(path.Dir(FileBasePath+"/"+filename), 0777)
		file, err := ioutil.TempFile("", "./result")
		if err != nil {
			log.Fatal("HTTPD ", err)
		}
		defer os.Remove(file.Name()) // clean up

		n, err := io.Copy(file, resp.Body)
		if err != nil {
			log.Error("NET ", err)
			delete(DownloadQueue, addr+filename)
			return
		}
		log.Debug("NET Received ", n, " bytes for file ", filename)

		err = os.Rename(file.Name(), FileBasePath+"/"+filename)
		if err != nil {
			log.Error("NET ", err)
			delete(DownloadQueue, addr+filename)
			return
		}
		UpdateFileIndex()
	} else {
		log.Debug("NET ", resp.StatusCode)
	}
	delete(DownloadQueue, addr+filename)
	log.Debug("NET Download removed from queue ", DownloadQueue)
}
