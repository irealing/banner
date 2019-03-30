package main

import (
	"github.com/qiniu/log"
	"os"
	"os/signal"
	"syscall"
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
	scheduler.Run()
}
