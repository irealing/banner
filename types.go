package main

import (
	"fmt"
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

type TaskDone func()
type Task struct {
	Host string
	Port uint
	Pro  Protocol
	Ack  TaskDone
}

func (t *Task) URL() string {
	return fmt.Sprintf("%s://%s:%d/", t.Pro, t.Host, t.Port)
}
