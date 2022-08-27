package main

import (
	"flag"
	"fmt"
	"github.com/inlets/connect/handler"
	"log"
	"net/http"
)

var (
	GitCommit string
	Version   string
)

func main() {
	var port int

	flag.IntVar(&port, "port", 3128, "The port to listen on")
	flag.Parse()

	log.Printf("Version: %s\tCommit: %s", Version, GitCommit)

	log.Printf("Listening on %d", port)

	http.ListenAndServe(fmt.Sprintf(":%d", port), handler.Handle())
}