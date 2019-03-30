package main

import "log"

func main() {
	cfg := DefaultConfig
	if scheduler, err := NewScheduler(cfg); err != nil {
		log.Fatal(err)
	} else {
		defer scheduler.Close()
		scheduler.Run()
	}
}
