package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-icap/icap"
)

var (
	ISTag = "\"GOLANG\""
	lines []string
)

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func main() {
	var err error
	lines, err = readLines("filter_list.txt")
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	icap.HandleFunc("/request", filter)
	icap.ListenAndServe(":13440", icap.HandlerFunc(filter))
}

func filter(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", ISTag)
	h.Set("Service", "Golang icap")

	switch req.Method {
	case "OPTIONS":
		h.Set("Methods", "REQMOD")
		h.Set("Allow", "204")
		h.Set("Preview", "0")
		h.Set("Transfer-Preview", "*")
		w.WriteHeader(200, nil, false)
	case "REQMOD":
		log.Printf("Got RequestURI : [%s] ", req.Request.RequestURI)
		if contains(lines, req.Request.RequestURI) {
			w.WriteHeader(200, &http.Response{
				StatusCode: http.StatusForbidden,
				Status:     "Forbidden",
			}, false)

		} else {
			w.WriteHeader(204, nil, false)
		}

	case "ERRDUMMY":
		w.WriteHeader(400, nil, false)
		fmt.Println("Malformed request")
	default:
		w.WriteHeader(405, nil, false)
		fmt.Println("Invalid request method")
	}
}
