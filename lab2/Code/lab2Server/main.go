package main

import (
	"lab2Server/cli"
	"lab2Server/server"
	"log"
)

// Well, due to my work i was in really big hurry, so sorry for code quality.

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	config := cli.ParseArgs()

	err := server.NewServer(config.Host, config.Port, config.Interval).Run()
	if err != nil {
		log.Fatal(err)
	}
}
