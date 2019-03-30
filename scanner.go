package main

import (
	"context"
	"log"
	"net/http"
	"io"
	"bytes"
	"regexp"
	"errors"
	"sync"
)

type Scanner struct {
	ctx    context.Context
	cancel context.CancelFunc
	task   <-chan *Task
	writer chan<- *Result
	client *http.Client
	wg     *sync.WaitGroup
}

func NewScanner(ctx context.Context, tc <-chan *Task, w chan<- *Result, client *http.Client, wg *sync.WaitGroup) *Scanner {
	c, cl := context.WithCancel(ctx)
	return &Scanner{ctx: c, cancel: cl, task: tc, writer: w, client: client, wg: wg}
}
func (scanner *Scanner) Run() {
loop:
	for {
		select {
		case <-scanner.ctx.Done():
			log.Print("context done")
			break loop
		case task, ok := <-scanner.task:
			if !ok {
				log.Println("task queue closed")
				break loop
			}
			log.Println("recv new task", task.Pro, task.Host, task.Port)
			if ret, err := scanner.capture(task); err != nil {
				continue loop
			} else {
				scanner.writer <- ret
			}
		}
	}
	scanner.wg.Done()
}
func (scanner *Scanner) capture(task *Task) (*Result, error) {
	resp, err := scanner.request(task.URL())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	server := resp.Header.Get("Server")
	title, err := scanner.getTitle(resp.Body)
	return &Result{Host: task.Host, Pro: task.Pro, Port: task.Port, Server: server, Title: title}, nil
}
func (scanner *Scanner) request(url string) (*http.Response, error) {
	log.Println("request url ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	return scanner.client.Do(req)
}

func (scanner *Scanner) getTitle(r io.Reader) (string, error) {
	buf := bytes.Buffer{}
	defer buf.Reset()
	buf.ReadFrom(r)
	re := regexp.MustCompile(regexTitle)
	value := re.FindString(buf.String())
	if value == emptyString {
		return value, errors.New("找不到结果")
	}
	return value[7: len(value)-8], nil
}
func (scanner *Scanner) Close() {
	scanner.cancel()
}
