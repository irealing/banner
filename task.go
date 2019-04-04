package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/qiniu/log"
)

type Protocol string

// Port 端口
type Port struct {
	Proto Protocol
	Port  uint
}

// portIterator Port迭代器
type portIterator struct {
	cursor int
	ports  []*Port
}

// NewPortGetter 返回已初始化的PortGetter
func NewPortGetter(filename string) (*PortGetter, error) {
	pg := &PortGetter{filename: filename}
	if err := pg.Init(); err != nil {
		return nil, err
	}
	return pg, nil
}

func (pi *portIterator) HasNext() bool {
	return pi.cursor < len(pi.ports)
}
func (pi *portIterator) Next() (p *Port) {
	p = pi.ports[pi.cursor]
	pi.cursor += 1
	return
}
func (pi *portIterator) Reset() error {
	if pi.ports == nil || len(pi.ports) == 0 {
		return errors.New("端口迭代器不可用")
	}
	pi.cursor = 0
	return nil
}

type PortGetter struct {
	filename string
	ports    []*Port
}

func (pg *PortGetter) Size() int {
	if pg.ports == nil {
		return 0
	}
	return len(pg.ports)
}

// Init 初始化
func (pg *PortGetter) Init() error {
	f, err := os.Open(pg.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	rec, err := reader.ReadAll()
	if err != nil {
		return err
	}
	size := len(rec)
	tmp := make([]*Port, size)
	cursor := 0
	for _, row := range rec {
		if len(row) != 2 {
			continue
		}
		hs, ps := row[0], row[1]
		port, _ := strconv.ParseInt(ps, 10, 32)
		if port > 65535 || port < 1 {
			continue
		}
		p := &Port{Proto: Protocol(hs), Port: uint(port)}
		tmp[cursor] = p
		cursor += 1
	}
	pg.ports = tmp[:cursor]
	return nil
}

// Iter 返回Port迭代器
func (pg *PortGetter) Iter() (*portIterator, error) {
	if pg.ports == nil || len(pg.ports) < 1 {
		return nil, errors.New("没有需要扫描的端口号")
	}
	return &portIterator{ports: pg.ports}, nil
}

type taskMaker struct {
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan *Task
	pg     *PortGetter
	input  io.ReadWriteCloser
	wg     *sync.WaitGroup
}

func newMaker(host, port string, cc uint, ctx context.Context) (*taskMaker, error) {
	pg, err := NewPortGetter(port)
	if err != nil {
		return nil, err
	}
	input, err := os.Open(host)
	if err != nil {
		return nil, err
	}
	c, cl := context.WithCancel(ctx)
	ch := make(chan *Task, cc)
	return &taskMaker{ctx: c, cancel: cl, ch: ch, pg: pg, input: input, wg: &sync.WaitGroup{}}, nil
}
func (tm *taskMaker) channel() <-chan *Task {
	return tm.ch
}
func (tm *taskMaker) Close() {
	tm.cancel()
	tm.input.Close()
	close(tm.ch)
}
func (tm *taskMaker) Run(callback func()) error {
	reader := bufio.NewReader(tm.input)
	it, _ := tm.pg.Iter()
loop:
	for {
		line, _, err := reader.ReadLine()
		host := string(line)
		if err != nil {
			log.Warn("taskMaker read file error", err)
			break
		}
		it.Reset()
		for it.HasNext() {
			select {
			case <-tm.ctx.Done():
				log.Debug("taskMaker context done")
				break loop
			default:
				p := it.Next()
				log.Debug("push new task ", p.Proto, host, p.Port)
				task := &Task{Host: host, Pro: p.Proto, Port: p.Port, Ack: tm}
				tm.Ready()
				tm.ch <- task
				if callback != nil {
					callback()
				}
			}
		}
	}
	tm.wg.Wait()
	log.Debug("all ready tasks ack")
	return nil
}
func (tm *taskMaker) Ready() {
	log.Debug("task ready")
	tm.wg.Add(1)
}
func (tm *taskMaker) Ack() {
	log.Debug("task done")
	tm.wg.Done()
}
