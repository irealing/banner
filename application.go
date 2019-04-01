package main

import (
	"context"
	"github.com/qiniu/log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Scheduler struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        *sync.WaitGroup
	writer    *os.File
	cfg       *AppConfig
	closeOnce sync.Once
}

func NewScheduler(cfg *AppConfig) (*Scheduler, error) {
	writer, err := os.Create(cfg.Output)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	return &Scheduler{cfg: cfg, ctx: ctx, cancel: cancel, wg: wg, writer: writer}, nil
}
func (scheduler *Scheduler) Run() error {
	cfg := scheduler.cfg
	tm, err := newMaker(cfg.Input, cfg.Port, scheduler.ctx)
	if err != nil {
		return err
	}
	defer tm.Close()
	saver := newTextSaver(scheduler.writer)
	defer saver.Close()
	var i uint
	for ; i < scheduler.cfg.Go; i++ {
		http.DefaultClient.Timeout = time.Duration(scheduler.cfg.TTL) * time.Second
		s := NewScanner(scheduler.ctx, tm.channel(), saver, http.DefaultClient, scheduler.wg)
		go s.Run()
	}
	go saver.Run()
	return tm.Run()
}
func (scheduler *Scheduler) Close() {
	scheduler.closeOnce.Do(scheduler.close)
}
func (scheduler *Scheduler) close() {
	log.Info("scheduler exit ...")
	scheduler.cancel()
	scheduler.wg.Wait()
	log.Debug("all scanner exit")
	if err := scheduler.writer.Close(); err != nil {
		log.Warn("failed to close file ", scheduler.writer.Name())
	}
}
