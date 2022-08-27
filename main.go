package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/inlets/connect/config"
	"github.com/inlets/connect/handler"
)

var (
	GitCommit string
	Version   string
)

func main() {
	var (
		port              int
		defaultPort       = 3128
		httpTimeout       = time.Second * 5
		maxHeaderBytes    = 1024 * 8
		configPath        string
		defaultConfigPath = "./config.yml"
	)

	flag.IntVar(&port, "port", defaultPort, "The port to listen on")
	flag.StringVar(&configPath, "config", defaultConfigPath, "Path to config file")
	flag.Parse()

	log.Printf("Version: %s\tCommit: %s", Version, GitCommit)

	log.Printf("Listening on %d", port)

	conf := config.Init(configPath)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: httpTimeout,
		Handler:           handler.Handle(),
		ReadTimeout:       httpTimeout,
		WriteTimeout:      httpTimeout,
		IdleTimeout:       httpTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		TLSNextProto:      nil,
		TLSConfig:         nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}

	handler.Init(conf)

	server.Handler = handler.Handle()
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
