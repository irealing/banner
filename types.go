package main

import (
	"context"
	"fmt"
	"github.com/qiniu/log"
	"os"
	"sync"
)

const (
	regexTitle = "(?i)<title>.*</title>"
	userAgent  = "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.23 Mobile Safari/537.36"
)

type Ack interface {
	Ready()
	Ack()
}

// Result 处理结果
type Result struct {
	Host   string
	Pro    Protocol
	Port   uint
	Server string
	Title  string
}

// String 转string
func (r *Result) String() string {
	return fmt.Sprintf("{Host: %s, Server: %s,Title: %s}", r.Host, r.Server, r.Title)
}

type Task struct {
	Host string
	Port uint
	Pro  Protocol
	Ack  Ack
}

func (t *Task) URL() string {
	return fmt.Sprintf("%s://%s:%d/", t.Pro, t.Host, t.Port)
}

type Saver interface {
	Save(result *Result)
	Run()
	Close()
}
type textSaver struct {
	wg     sync.WaitGroup
	file   *os.File
	ctx    context.Context
	cancel context.CancelFunc
	wc     chan *Result
}

func (ts *textSaver) Run() {
loop:
	for {
		select {
		case <-ts.ctx.Done():
			break loop
		case r, ok := <-ts.wc:
			if !ok {
				break loop
			}
			s := fmt.Sprintf("%s,%d,%s,%s", r.Host, r.Port, r.Server, r.Title)
			if _, err := ts.file.WriteString(s); err != nil {
				log.Warn("failed to write string ", s)
			}
			ts.wg.Done()
			log.Debug("saver add -1")
		}
	}
	ts.wg.Wait()
}

func (ts *textSaver) Save(result *Result) {
	ts.wc <- result
	ts.wg.Add(1)
	log.Debug("saver result ", result)
}

func (ts *textSaver) Close() {
	ts.wg.Wait()
	ts.cancel()
	close(ts.wc)
}
func newTextSaver(writer *os.File) Saver {
	ctx, cancel := context.WithCancel(context.Background())
	return &textSaver{file: writer, ctx: ctx, cancel: cancel, wc: make(chan *Result)}
}
