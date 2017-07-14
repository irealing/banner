package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/qiniu/log"
)

const (
	regexTitle = "(?i)<title>.*</title>"
	userAgent  = "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.23 Mobile Safari/537.36"
)

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
}

func (t *Task) URL() string {
	return fmt.Sprintf("%s://%s:%d/", t.Pro, t.Host, t.Port)
}

// Capturer 获取主机Server及title
type Capturer struct {
	client   *http.Client
	protocol string
}

// Capture 获取信息
func (c *Capturer) Capture(t *Task) (*Result, error) {
	resp, err := c.request(t.URL())
	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}
	log.Info("请求成功: ", t.URL())
	defer resp.Body.Close()
	server := resp.Header.Get("Server")
	title, err := c.getTitle(resp.Body)
	return &Result{Host: t.Host, Pro: t.Pro, Port: t.Port, Server: server, Title: title}, nil
}
func (c *Capturer) getTitle(r io.Reader) (string, error) {
	buf := bytes.Buffer{}
	defer buf.Reset()
	buf.ReadFrom(r)
	re := regexp.MustCompile(regexTitle)
	value := re.FindString(buf.String())
	if value == "" {
		return value, errors.New("找不到结果")
	}
	return value[7: len(value)-8], nil
}

func (c *Capturer) request(url string) (*http.Response, error) {
	client := c.client
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	return client.Do(req)
}

// NewCapturer 获取Capturer对象
func NewCapturer() *Capturer {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Timeout: 10 * time.Second, Transport: transport}
	return &Capturer{client: client}
}

// Scheduler 调度器
type Scheduler struct {
	wChan    chan *Result
	rChan    chan *Task
	reader   io.Reader
	writer   io.Writer
	counter  uint
	con      uint
	capturer *Capturer
}

//NewScheduler 创建Scheduler
func NewScheduler(con uint, r io.Reader, w io.Writer) *Scheduler {
	c := NewCapturer()
	scheduler := &Scheduler{
		wChan:    make(chan *Result, con),
		rChan:    make(chan *Task, con),
		con:      con,
		reader:   r,
		writer:   w,
		capturer: c,
	}
	return scheduler
}

// Run 运行程序逻辑
func (s *Scheduler) Run() uint {
	var i uint
	for i = 0; i < s.con; i++ {
		go s.work(i)
	}
	countChan := make(chan uint)
	go s.makeTask(countChan)
	var waited uint
	over := false
	for {
		select {
		case <-countChan:
			log.Debug("无更多任务")
			over = true
		case <-time.After(time.Millisecond):
			over = false
		}
		if over {
			log.Debugf("读取剩余结果 %d-%d", s.counter, waited)
			s.wait(s.counter - waited)
			break
		}
		s.wait(1)
		waited += 1
		log.Debug("waitd ", waited)
	}
	return s.counter
}
func (s *Scheduler) makeTask(count chan uint) {
	reader := bufio.NewReader(s.reader)
	iter, _ := NewPortGetter().Iter()
	var n uint
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Debug("文件读取结束")
			close(s.rChan)
			s.counter = n
			count <- n
			break
		}
		host := string(line)
		iter.Reset()
		for iter.HasNext() {
			n += 1
			p := iter.Next()
			s.rChan <- &Task{Host: host, Pro: p.Prot, Port: p.Port}
		}
	}
}
func (s *Scheduler) work(id uint) {
	log.Debug("启动goroutine")
	for {
		host, ok := <-s.rChan
		if !ok {
			log.Debug("goroutine 结束", id)
			break
		}
		log.Debugf("goroutine %d 收到任务：%s", id, host.URL())
		r, err := s.capturer.Capture(host)
		if err != nil {
			log.Warn("failed to load ", host.URL())
		}
		s.wChan <- r
	}
}
func (s *Scheduler) wait(n uint) {
	var x uint
	for ; x < n; x++ {
		r := <-s.wChan
		if r == nil {
			continue
		}
		log.Debug(r.String())
		str := fmt.Sprintf("%s,%d,%s,%s\n", r.Host, r.Port, r.Server, r.Title)
		s.writer.Write([]byte(str))
	}
}
