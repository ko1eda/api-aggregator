package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/ko1eda/apiaggregator/http"
	"github.com/ko1eda/apiaggregator/http/providers/koalaJsonEatery"
	"github.com/ko1eda/apiaggregator/http/providers/koalaXmlGrill"
)

func main() {
	port := flag.String("p", "8080", "Set the port the server will run on")

	flag.Parse()
	// create a new server and a new client for our providers to connect to their data sources
	// pass in the port dyanamically with config flag and variadic argument
	srvr := http.NewServer(http.WithAddress(":" + *port))

	client := http.NewClient()

	p := koalaJsonEatery.NewProvider(client)
	p2 := koalaXmlGrill.NewProvider(client)

	// Assign Providers to server
	srvr.JsonEateryProvider = p
	srvr.XmlGrillProvider = p2

	// open the sever (this is non blocking because we're running listener as goroutine)
	srvr.Open()

	// close the server when our main function exits
	defer srvr.Close()

	sigchan := make(chan os.Signal, 1)

	signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGQUIT)

	// block indefinitely for our server until we get an os system call
	<-sigchan
}
