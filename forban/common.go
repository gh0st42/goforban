package forban

import (
	"os"
	"time"

	uuid "github.com/gofrs/uuid"
)

type ForbanNode struct {
	name     string
	uuid     string
	hmac     string
	ipv4     string
	ipv6     string
	lastSeen time.Time
}

type ForbanNodeEntry struct {
	node       ForbanNode
	firstSeen  time.Time
	files      []FileEntry
	totalStore int64
}

// InitForban Initialize all relevant variables and update file index
func InitForban() {
	u1 := uuid.Must(uuid.NewV4())
	//u1 := uuid.NewV4()
	MyUuid = u1.String()
	thishost, _ := os.Hostname()
	MyName = "goforban@" + thishost
	MyPsk = "forban"
	Neighborhood = make(map[string]ForbanNodeEntry)
	HmacIgnoreList = make(map[string]int)
	DownloadQueue = make(map[string]bool)

	os.MkdirAll(FileBasePath+"/forban", 0755)

	UpdateFileIndex()
	currentHmac = GetIndexHmac()
}
