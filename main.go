package main

import (
	"flag"
	"fmt"
	"github.com/BruteForceFencer/core/config"
	"github.com/BruteForceFencer/core/dashboard"
	"github.com/BruteForceFencer/core/hitcounter"
	"github.com/BruteForceFencer/core/message-server"
	"github.com/BruteForceFencer/core/version"
	"os"
	"os/signal"
	"runtime"
)

var (
	Configuration *config.Configuration
	Dashboard     *dashboard.Server
	HitCounter    *hitcounter.HitCounter
	Server        *server.Server
)

func configure() {
	// Setup multithreading
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parse flags
	configFilename := flag.String("c", "config.json", "the name of the configuration file")
	displayVersion := flag.Bool("version", false, "display the version number")
	flag.Parse()

	// Display version number
	if *displayVersion {
		version.PrintVersion()
		os.Exit(0)
	}

	// Read the configuration
	var errs []error
	Configuration, errs = config.ReadConfig(*configFilename)
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, "configuration error:", err)
		}

		os.Exit(1)
	}
}

func initialize() {
	// Create the hit counter
	HitCounter = hitcounter.NewHitCounter(
		Configuration.Directions,
		Configuration.Logger,
	)

	// Create the server
	Server = new(server.Server)
	Server.HandleFunc = routeRequest

	// Create the dashboard
	if Configuration.DashboardAddress != "" {
		Dashboard = dashboard.New(Configuration, HitCounter)
	}
}

func start() {
	go Server.ListenAndServe(
		Configuration.ListenType,
		Configuration.ListenAddress,
	)

	if Dashboard != nil {
		go Dashboard.ListenAndServe()
	}
}

func routeRequest(req *server.Request) bool {
	return HitCounter.HandleRequest(req.Direction, req.Value)
}

func main() {
	configure()
	initialize()
	start()

	fmt.Fprintln(os.Stderr, "The server is running.")

	// Capture interrupt signal so that the server closes properly
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	<-interrupts

	Server.Close()
}
