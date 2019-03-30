package main

import (
	"context"
	"sync"
	"os"
	"net/http"
	"fmt"
	"log"
	"time"
)

type Scheduler struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        *sync.WaitGroup
	tm        *taskMaker
	writer    *os.File
	taskChan  <-chan *Task
	writeChan chan *Result
	cfg       *AppConfig
}

func NewScheduler(cfg *AppConfig) (*Scheduler, error) {
	writer, err := os.Create(cfg.Output)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	writeChan := make(chan *Result, cfg.Go)
	tm, err := newMaker(cfg.Input, cfg.Port, ctx)
	if err != nil {
		return nil, err
	}
	wg := &sync.WaitGroup{}
	return &Scheduler{cfg: cfg, ctx: ctx, cancel: cancel, wg: wg, tm: tm, writer: writer, taskChan: tm.channel(), writeChan: writeChan}, nil
}
func (scheduler *Scheduler) Run() {
	var i uint
	for ; i < scheduler.cfg.Go; i++ {
		http.DefaultClient.Timeout = time.Duration(scheduler.cfg.TTL) * time.Second
		s := NewScanner(scheduler.ctx, scheduler.taskChan, scheduler.writeChan, http.DefaultClient, scheduler.wg)
		scheduler.wg.Add(1)
		go s.Run()
	}
	go scheduler.tm.Run()
	scheduler.save()
}
func (scheduler *Scheduler) save() {
loop:
	for {
		select {
		case <-scheduler.ctx.Done():
			break loop
		case r, ok := <-scheduler.writeChan:
			if ok && r != nil {
				log.Print("recv result ", r.String())
				scheduler.writer.WriteString(fmt.Sprintf("%s,%d,%s,%s", r.Host, r.Port, r.Server, r.Title))
			} else if !ok {
				break loop
			}
		}
	}
}
func (scheduler *Scheduler) Close() {
	scheduler.cancel()
	scheduler.tm.Close()
	scheduler.wg.Wait()
	close(scheduler.writeChan)
	scheduler.writer.Close()
}
