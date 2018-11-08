package forban

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gh0st42/goforban/static"
	log "github.com/sirupsen/logrus"
)

func requestFilter(remoteaddr string, filename string) bool {

	return true
}

func uploadFilter(remoteaddr string, hash string, tempfile string) bool {
	return true
}

func handleServe(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.Form)
	//fmt.Println(r.RequestURI)
	//fmt.Println(r.FormValue("g"))
	if r.FormValue("g") != "" {
		var filename string
		if r.FormValue("f") == "b64e" {
			fname, err := base64.StdEncoding.DecodeString(strings.Replace(r.FormValue("g"), "!", "=", -1))
			if err != nil {
				log.Error("HTTPD ", err)
			}
			filename = string(fname)
		} else {
			filename = r.FormValue("g")
		}
		if requestFilter(r.RemoteAddr, filename) {
			localPath := FileBasePath + "/" + filename
			_, justfile := filepath.Split(localPath)
			w.Header().Set("Content-Disposition", "True; filename="+justfile)
			http.ServeFile(w, r, localPath)
			log.Info("HTTPD Serving "+localPath+" to ", r.RemoteAddr)
		} else {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "403")
			log.Warn("HTTPD 403 - "+r.URL.Path[1:], r.RemoteAddr)
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	localPath := FileBasePath + "/" + r.URL.Path[1:]
	log.Info("HTTPD Access to "+r.URL.Path[1:]+" "+localPath+" from ", r.RemoteAddr)
	//fmt.Fprintf(w, getindexhtml())

	if r.URL.Path[1:] == "index.html" || r.URL.Path[1:] == "" {
		fmt.Fprintf(w, getindexhtml())
	} else {
		if _, err := os.Stat(localPath); err == nil {
			log.Debug("HTTPD Delivering "+localPath+" to ", r.RemoteAddr)
			_, justfile := filepath.Split(localPath)
			//w.Header().Set("Content-Disposition", "True; filename=index")
			w.Header().Set("Content-Disposition", "True; filename="+justfile)
			http.ServeFile(w, r, localPath)
		} else {
			//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404")
			log.Warn("HTTPD 404 - "+r.URL.Path[1:], r.RemoteAddr)
		}
	}
}

// not really needed any more
func peerHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("HTTPD peers requested by ", r.RemoteAddr)
	fmt.Fprintf(w, "uuid                              \t name           \t hmac                 \t first seen \t last seen\n")
	for key, value := range Neighborhood {
		fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s\n",
			key, value.node.name, value.node.hmac,
			time.Since(value.firstSeen), time.Since(value.node.lastSeen))
	}
}

// local daemon control
func ctrlHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("HTTPD ctrl request to "+r.URL.Path[1:]+" from ", r.RemoteAddr)
	if !strings.HasPrefix(r.RemoteAddr, "[::1]:") && !strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Access Forbidden")
		log.Warn("HTTPD Illegal control access "+r.URL.Path[1:], " from ", r.RemoteAddr)
	}
	if strings.HasSuffix(r.URL.Path, "/stop") {
		log.Info("HTTPD Server shutdown requested")
		w.WriteHeader(http.StatusAccepted)
		go func() {
			time.Sleep(1 * time.Second)
			os.Exit(0)
		}()
	}
}

// Upload files to forban store
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.TempFile("", "./result")
	if err != nil {
		log.Fatal("NET ", err)
	}
	defer os.Remove(file.Name()) // clean up

	log.Info("NET Upload ", r)
	//fmt.Fprintf(w, "%v", r)
	n, err := io.Copy(file, r.Body)
	if err != nil {
		log.Fatal("NET ", err)
	}

	hasher := sha256.New()
	b, err := ioutil.ReadFile(file.Name())
	CheckError(err)
	hasher.Write(b)
	shasum := hex.EncodeToString(hasher.Sum(nil))
	log.Debug("NET Upload hash: ", shasum)
	if uploadFilter(r.RemoteAddr, shasum, file.Name()) {
		err = ioutil.WriteFile(FileBasePath+"/"+shasum, b, 0644)
		CheckError(err)
		log.Info("HTTPD Added ", shasum, " with ", len(b), " bytes")
		UpdateFileIndex()
	}
	w.Write([]byte(fmt.Sprintf("%d bytes are recieved.\n", n)))
}

// ServeHttpd starts serving forban content
func ServeHttpd() {

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(static.HTTP)))
	http.HandleFunc("/ctrl/", ctrlHandler)
	http.HandleFunc("/", handler)
	http.HandleFunc("/s/", handleServe)
	http.HandleFunc("/peers", peerHandler) // deprecated
	http.HandleFunc("/upload", uploadHandler)

	http.Handle("/bundles/", http.StripPrefix("/bundles/", http.FileServer(http.Dir(FileBasePath))))
	//http.Handle("/assets", static.Handler)
	http.ListenAndServe(":12555", nil)
}
