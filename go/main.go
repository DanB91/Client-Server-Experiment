package main

import (
	"log"
	"net/http"
	"os"
)

type RequestContext struct {
	index_html_file []byte
}

func (rc *RequestContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write(rc.index_html_file)
}

func main() {
	if index_html_file, err := os.ReadFile("../index.html"); err == nil {
		rc := &RequestContext{index_html_file: index_html_file}
		http.Handle("/", rc)
		log.Fatal(http.ListenAndServe("", nil))
	} else {
		fatal("Error starting up server %v", err)
	}

}

func println(f string, args ...any) {
	log.Printf(f+"\n", args...)
}

func fatal(f string, args ...any) {
	log.Fatalf(f, args...)
}
