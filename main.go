package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/qiniu/log"
)

func main() {
	cfg := DefaultConfig
	scheduler, err := NewScheduler(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer scheduler.Close()
	go func(sch *Scheduler) {
		sign := make(chan os.Signal)
		signal.Notify(sign, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
		log.Warn("Receive signal", <-sign)
		scheduler.Close()
	}(scheduler)
	err = scheduler.Run()
	if err != nil {
		log.Warn("scheduler run error", err)
	}
}
