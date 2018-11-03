package forban

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gh0st42/goforban/static"
	log "github.com/sirupsen/logrus"
)

func handleServe(w http.ResponseWriter, r *http.Request) {

	//fmt.Println(r.Form)
	//fmt.Println(r.RequestURI)
	//fmt.Println(r.FormValue("g"))
	if r.FormValue("g") != "" {
		var filename string
		if r.FormValue("f") == "b64e" {
			fname, err := base64.StdEncoding.DecodeString(strings.Replace(r.FormValue("g"), "!", "=", -1))
			if err != nil {
				log.Error(err)
			}
			filename = string(fname)
		} else {
			filename = r.FormValue("g")
		}
		localPath := FileBasePath + "/" + filename
		_, justfile := filepath.Split(localPath)
		w.Header().Set("Content-Disposition", "True; filename="+justfile)
		http.ServeFile(w, r, localPath)
		log.Info("serving " + localPath)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	localPath := FileBasePath + "/" + r.URL.Path[1:]
	log.Info("Access to " + r.URL.Path[1:] + " " + localPath)
	//fmt.Fprintf(w, getindexhtml())

	if r.URL.Path[1:] == "index.html" || r.URL.Path[1:] == "" {
		fmt.Fprintf(w, getindexhtml())
	} else {
		if _, err := os.Stat(localPath); err == nil {
			log.Debug("Delivering " + localPath)
			_, justfile := filepath.Split(localPath)
			//w.Header().Set("Content-Disposition", "True; filename=index")
			w.Header().Set("Content-Disposition", "True; filename="+justfile)
			http.ServeFile(w, r, localPath)
		} else {
			//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404")
			log.Warn("404 " + r.URL.Path[1:])
		}
	}
}

// not really needed any more
func peerHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("peers requested")
	fmt.Fprintf(w, "uuid                              \t name           \t hmac                 \t first seen \t last seen\n")
	for key, value := range Neighborhood {
		fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s\n",
			key, value.node.name, value.node.hmac,
			time.Since(value.firstSeen), time.Since(value.node.lastSeen))
	}
}

// ServeHttpd starts serving forban content
func ServeHttpd() {

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(static.HTTP)))
	http.HandleFunc("/", handler)
	http.HandleFunc("/s/", handleServe)

	http.HandleFunc("/peers", peerHandler) // deprecated
	//http.Handle("/assets", static.Handler)
	http.ListenAndServe(":12555", nil)
}
