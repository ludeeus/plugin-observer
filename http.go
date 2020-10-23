package main

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

func checkNetwork(r *http.Request) bool {
	remote := strings.Split(r.RemoteAddr, ":")[0]
	remoteIP := net.ParseIP(remote)

	// Is in network
	if !hassioNetwork.Contains(remoteIP) {
		log.Printf("Access not allow from %s", remote)
		return false
	}

	return true
}

func apiPing(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func apiLogs(w http.ResponseWriter, r *http.Request) {
	if !checkNetwork(r) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	log.Printf("Access to logs from %s", r.RemoteAddr)

	err := supervisorLogs(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the content
	w.Header().Add("Content-Type", "text/plain")
}

func apiRestart(w http.ResponseWriter, r *http.Request) {
	if !checkNetwork(r) || supervisorPing() {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	log.Printf("Access to restart from %s", r.RemoteAddr)

	err := supervisorRestart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type statusData struct {
	On   bool
	Logs string
}

func statusIndex(w http.ResponseWriter, r *http.Request) {
	data := statusData{
		On:   supervisorPing(),
		Logs: "",
	}

	// Set logs
	if !data.On {
		var buf bytes.Buffer
		var re = regexp.MustCompile(`\[\d+m`)
		logWriter := bufio.NewWriter(&buf)

		err := supervisorLogs(logWriter)
		if err != nil {
			data.Logs = err.Error()
		} else {
			data.Logs = re.ReplaceAllLiteralString(buf.String(), "")
		}
	}

	// Render Website
	indexTemplate.Execute(w, data)
}
