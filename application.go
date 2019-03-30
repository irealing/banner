package main

import (
	"context"
	"sync"
	"os"
	"net/http"
	"fmt"
	"github.com/qiniu/log"
	"time"
)

type Scheduler struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        *sync.WaitGroup
	writer    *os.File
	writeChan chan *Result
	cfg       *AppConfig
	closeOnce sync.Once
}

func NewScheduler(cfg *AppConfig) (*Scheduler, error) {
	writer, err := os.Create(cfg.Output)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	writeChan := make(chan *Result, cfg.Go)
	wg := &sync.WaitGroup{}
	return &Scheduler{cfg: cfg, ctx: ctx, cancel: cancel, wg: wg, writer: writer, writeChan: writeChan}, nil
}
func (scheduler *Scheduler) Run() error {
	cfg := scheduler.cfg
	tm, err := newMaker(cfg.Input, cfg.Port, scheduler.ctx)
	if err != nil {
		return err
	}
	defer tm.Close()
	var i uint
	for ; i < scheduler.cfg.Go; i++ {
		http.DefaultClient.Timeout = time.Duration(scheduler.cfg.TTL) * time.Second
		s := NewScanner(scheduler.ctx, tm.channel(), scheduler.writeChan, http.DefaultClient, scheduler.wg)
		scheduler.wg.Add(1)
		go s.Run()
	}
	go scheduler.save()
	return tm.Run()
}
func (scheduler *Scheduler) save() {
loop:
	for {
		select {
		case <-scheduler.ctx.Done():
			break loop
		case r, ok := <-scheduler.writeChan:
			if ok && r != nil {
				log.Info("recv result ", r.String())
				scheduler.writer.WriteString(fmt.Sprintf("%s,%d,%s,%s\n", r.Host, r.Port, r.Server, r.Title))
			} else if !ok {
				break loop
			}
		}
	}
}
func (scheduler *Scheduler) Close() {
	scheduler.closeOnce.Do(scheduler.close)
}
func (scheduler *Scheduler) close() {
	log.Info("scheduler exit ...")
	scheduler.cancel()
	scheduler.wg.Wait()
	log.Debug("all scanner exit")
	close(scheduler.writeChan)
	scheduler.writer.Close()
}
