package main

import (
	"context"
	"crypto/tls"
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

func (scheduler *Scheduler) Ready() {
	scheduler.wg.Add(1)
}

func (scheduler *Scheduler) Ack() {
	scheduler.wg.Done()
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
	tm, err := newMaker(cfg.Input, cfg.Port, scheduler.cfg.Go, scheduler.ctx)
	if err != nil {
		return err
	}
	saver := newTextSaver(scheduler.writer, cfg.Go)
	defer saver.Close()
	defer scheduler.wg.Wait()
	defer tm.Close()
	http.DefaultClient.Timeout = time.Duration(scheduler.cfg.TTL) * time.Second
	http.DefaultTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	http.DefaultClient.Transport = http.DefaultTransport
	go saver.Run()
	return tm.Run(scheduler.startGo(tm, saver))
}
func (scheduler *Scheduler) startGo(tm *taskMaker, saver Saver) func() {
	var i uint
	return func() {
		if i >= scheduler.cfg.Go {
			return
		}
		i += 1
		log.Debug("start goroutine", i)
		s := NewScanner(tm.channel(), saver, http.DefaultClient, scheduler)
		scheduler.Ready()
		go s.Run()
	}
}
func (scheduler *Scheduler) Close() {
	scheduler.closeOnce.Do(scheduler.close)
}
func (scheduler *Scheduler) close() {
	log.Info("scheduler exit ...")
	scheduler.cancel()
	log.Debug("all scanner exit")
	if err := scheduler.writer.Close(); err != nil {
		log.Warn("failed to close file ", scheduler.writer.Name())
	}
}
