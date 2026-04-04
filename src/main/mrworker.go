package main

//
// start a worker process, which is implemented
// in ../mr/worker.go. typically there will be
// multiple worker processes, talking to one master.
//
// go run mrworker.go wc.so
//
// Please do not change this file.
//

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"plugin"

	"../mr"
)

type Config struct {
	Master string `json:"master"`
}

func loadConfig(filename string) Config {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Warning: cannot read config file %s, using default 127.0.0.1:1234\n", filename)
		return Config{Master: "127.0.0.1:1234"}
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Warning: config file format error, using default\n")
		return Config{Master: "127.0.0.1:1234"}
	}

	if config.Master == "" {
		config.Master = "127.0.0.1:1234"
	}

	fmt.Printf("Master address from config: %s\n", config.Master)
	return config
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrworker xxx.so\n")
		os.Exit(1)
	}

	config := loadConfig("config.json")

	mapf, reducef := loadPlugin(os.Args[1])

	mr.Worker(config.Master, mapf, reducef)
}

func loadPlugin(filename string) (func(string, string) []mr.KeyValue, func(string, []string) string) {
	p, err := plugin.Open(filename)
	if err != nil {
		log.Fatalf("cannot load plugin %v", filename)
	}
	xmapf, err := p.Lookup("Map")
	if err != nil {
		log.Fatalf("cannot find Map in %v", filename)
	}
	mapf := xmapf.(func(string, string) []mr.KeyValue)
	xreducef, err := p.Lookup("Reduce")
	if err != nil {
		log.Fatalf("cannot find Reduce in %v", filename)
	}
	reducef := xreducef.(func(string, []string) string)

	return mapf, reducef
}
