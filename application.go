package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/qiniu/log"
)

type Scheduler struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        *sync.WaitGroup
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
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	return &Scheduler{cfg: cfg, ctx: ctx, cancel: cancel, wg: wg}, nil
}
func (scheduler *Scheduler) Run() error {
	cfg := scheduler.cfg
	tm, err := newMaker(cfg.Input, cfg.Port, scheduler.cfg.Go, scheduler.ctx)
	if err != nil {
		return err
	}
	file, err := os.Create(scheduler.cfg.Output)
	if err != nil {
		log.Warn("failed to create output file ", err)
		return err
	}
	defer file.Close()
	saver := newTextSaver(file, cfg.Go)
	defer saver.Close()
	defer scheduler.wg.Wait()
	defer tm.Close()
	scheduler.makeHttpClient()
	go saver.Run()
	return tm.Run(scheduler.startGo(tm, saver))
}
func (scheduler *Scheduler) makeHttpClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transport, Timeout: time.Duration(scheduler.cfg.TTL) * time.Second}
	return client
}
func (scheduler *Scheduler) startGo(tm *taskMaker, saver Saver) func() {
	var i uint
	client := scheduler.makeHttpClient()
	return func() {
		if i >= scheduler.cfg.Go {
			return
		}
		i += 1
		log.Debug("start goroutine", i)
		s := NewScanner(tm.channel(), saver, client, scheduler)
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
}
