package forban

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func Stop() {
	res, err := http.Get("http://127.0.0.1:12555/ctrl/stop")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
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
