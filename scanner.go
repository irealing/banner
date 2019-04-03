package main

import (
	"bytes"
	"errors"
	"github.com/qiniu/log"
	"io"
	"net/http"
	"regexp"
	"sync/atomic"
)

var globalScannerId int32

type Scanner struct {
	id     int32
	task   <-chan *Task
	saver  Saver
	client *http.Client
	ack    Ack
}

func NewScanner(tc <-chan *Task, saver Saver, client *http.Client, ack Ack) *Scanner {
	id := atomic.AddInt32(&globalScannerId, 1)
	return &Scanner{id: id, task: tc, saver: saver, client: client, ack: ack}
}
func (scanner *Scanner) ID() int32 {
	return scanner.id
}
func (scanner *Scanner) Run() {
	defer scanner.ack.Ack()
	c := 0
loop:
	for {
		select {
		case task, ok := <-scanner.task:
			if !ok {
				log.Info("task queue closed scanner ", scanner.id)
				break loop
			}
			log.Debug("recv new task", task.Pro, task.Host, task.Port, scanner.id)
			//task.Ack.Ready()
			ret, err := scanner.capture(task)
			task.Ack.Ack()
			c++
			if err != nil {
				log.Infof("scanner %d execute task %d-%d failed", scanner.id, scanner.id, c)
				continue loop
			} else {
				log.Infof("scanner %d execute task %d-%d success %s", scanner.id, scanner.id, c, ret.String())
				scanner.saver.Save(ret)
			}
		}
	}
	log.Debug("scanner done", scanner.id)
}
func (scanner *Scanner) capture(task *Task) (*Result, error) {
	resp, err := scanner.request(task.URL())
	if err != nil {
		log.Debug("failed to request ", task.URL(), err)
		return nil, err
	}
	defer resp.Body.Close()
	server := resp.Header.Get("Server")
	title, err := scanner.getTitle(resp.Body)
	return &Result{Host: task.Host, Pro: task.Pro, Port: task.Port, Server: server, Title: title}, nil
}
func (scanner *Scanner) request(url string) (*http.Response, error) {
	log.Debugf("scanner %d request url %s", scanner.id, url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Debug("failed to load", err)
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
	return value[7 : len(value)-8], nil
}
