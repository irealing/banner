package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/qiniu/log"
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
	wg   sync.WaitGroup
	file *os.File
	ctx  context.Context
	wc   chan *Result
}

func (ts *textSaver) Run() {
	writer := csv.NewWriter(ts.file)
	defer ts.wg.Done()
	for {
		r, ok := <-ts.wc
		if !ok {
			log.Debug("textSaver break")
			break
		}
		line := []string{r.Host, strconv.FormatUint(uint64(r.Port), 10), r.Server, r.Title}
		if err := writer.Write(line); err != nil {
			log.Warn("failed to write csv ", err)
		}
		ts.wg.Done()
		log.Debug("saver add -1")
	}
	writer.Flush()
}

func (ts *textSaver) Save(result *Result) {
	ts.wc <- result
	ts.wg.Add(1)
	log.Debug("saver result ", result)
}

func (ts *textSaver) Close() {
	ts.wg.Wait()
	close(ts.wc)
	ts.wg.Add(1)
	ts.wg.Wait()
}
func newTextSaver(writer *os.File, cache uint) Saver {
	return &textSaver{file: writer, wc: make(chan *Result, cache)}
}
