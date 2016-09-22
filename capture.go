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

	"runtime"

	"github.com/qiniu/log"
)

const regex_title = "(?i)<title>.*</title>"

// Result 处理结果
type Result struct {
	Host   string
	Server string
	Title  string
}

// String 转string
func (r *Result) String() string {
	return fmt.Sprintf("{Host: %s, Server: %s,Title: %s}", r.Host, r.Server, r.Title)
}

// Capturer 获取主机Server及title
type Capturer struct {
	client   *http.Client
	protocol string
}

// Capture 获取信息
func (c *Capturer) Capture(host string) (*Result, error) {
	resp, err := c.request(host)
	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}
	log.Info("请求成功: ", host)
	defer resp.Body.Close()
	server := resp.Header.Get("Server")
	title, err := c.getTitle(resp.Body)
	return &Result{Host: host, Server: server, Title: title}, nil
}
func (c *Capturer) getTitle(r io.Reader) (string, error) {
	buf := bytes.Buffer{}
	defer buf.Reset()
	buf.ReadFrom(r)
	re := regexp.MustCompile(regex_title)
	value := re.FindString(buf.String())
	if value == "" {
		return value, errors.New("找不到结果")
	}
	return value[7 : len(value)-8], nil
}

func (c *Capturer) request(host string) (*http.Response, error) {
	urlPrefix := map[string]string{"http": "http://", "https": "https://"}
	client := c.client
	url := fmt.Sprintf("%s%s", urlPrefix[c.protocol], host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.23 Mobile Safari/537.36")
	return client.Do(req)
}

// NewCapturer 获取Capturer对象
func NewCapturer(protocol string) *Capturer {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Timeout: 5 * time.Second, Transport: transport}
	return &Capturer{client: client, protocol: protocol}
}

// Scheduler 调度器
type Scheduler struct {
	wChan    chan *Result
	rChan    chan string
	reader   io.Reader
	writer   io.Writer
	counter  uint
	con      uint
	curCon   uint
	capturer *Capturer
}

//NewScheduler 创建Schema
func NewScheduler(con uint, r io.Reader, w io.Writer, proto string) *Scheduler {
	capturer := NewCapturer(proto)
	scheduler := &Scheduler{wChan: make(chan *Result, con), rChan: make(chan string, con), con: con, reader: r, writer: w, capturer: capturer}
	return scheduler
}

// Run 运行程序逻辑
func (s *Scheduler) Run() uint {
	var i uint
	for i = 0; i < s.con; i++ {
		go s.start(i)
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
		log.Debug("waited ", waited)
	}
	return s.counter
}
func (s *Scheduler) makeTask(count chan uint) {
	reader := bufio.NewReader(s.reader)
	var n uint
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Debug("文件读取结束")
			close(s.rChan)
			count <- n
			break
		}
		n += 1
		host := string(line)
		s.rChan <- string(host)
	}
	s.counter = n
}
func (s *Scheduler) start(id uint) {
	s.curCon += 13
	log.Debug("启动goroutine")
	for {
		host, ok := <-s.rChan
		if !ok {
			log.Debug("goroutine 退出", id)
			break
		}
		log.Debugf("goroutine %d 收到任务：%s", id, host)
		r, err := s.capturer.Capture(host)
		if err != nil {
			log.Warn("failed to load ", host)
		}
		s.wChan <- r
	}
	runtime.Goexit()
}
func (s *Scheduler) wait(n uint) {
	log.Debug("waite", n)
	var x uint
	for ; x < n; x++ {
		r := <-s.wChan
		if r == nil {
			continue
		}
		log.Debug(r.String())
		str := fmt.Sprintf("%s,%s,%s\n", r.Host, r.Server, r.Title)
		s.writer.Write([]byte(str))
	}
}
