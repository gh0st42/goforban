package forban

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

// CheckError Generic error checking, exit in case of error
func CheckError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func GetIndexHmac() string {
	//dat, err := ioutil.ReadFile("/home/gh0st/LocalCode/Forban/var/share/forban/index")
	dat, err := ioutil.ReadFile(FileBasePath + "/forban/index")
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(dat))

	idxHmac := hmac.New(sha1.New, []byte(MyPsk))
	idxHmac.Write(dat)
	currentHmac = hex.EncodeToString(idxHmac.Sum(nil))
	return hex.EncodeToString(idxHmac.Sum(nil))
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetTotalFileSize(node ForbanNodeEntry) int64 {
	var count int64
	count = 0
	for _, i := range node.files {
		count += i.Size
	}
	return count
}
