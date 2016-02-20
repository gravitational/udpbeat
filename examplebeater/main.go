package main

import (
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/trace"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&trace.TextFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
	h, err := trace.NewUDPHook()
	if err != nil {
		log.Fatalf("hook: %v", err)
	}
	log.AddHook(h)

	go func() {
		log.Infof("got it!")
		for {
			log.Infof("something new")
			e := log.WithFields(log.Fields{trace.Component: "play"})
			e.Infof("this time")
			time.Sleep(time.Second)
		}
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, os.Kill)

	<-s
}
