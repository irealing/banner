package main

import (
	"errors"
	"os"
	"encoding/csv"
	"strconv"
)

type Protocol string

const (
	HTTPProtocol  Protocol = "http"
	HTTPSProtocal Protocol = "https"
	portFile               = "ports.csv"
)

// Port 端口
type Port struct {
	Prot Protocol
	Port uint
}

// PortIter Port迭代器
type PortIter struct {
	cursor int
	ports  []*Port
}

func (pi *PortIter) HasNext() bool {
	return pi.cursor < len(pi.ports)
}
func (pi *PortIter) Next() (p *Port) {
	p = pi.ports[pi.cursor]
	pi.cursor += 1
	return
}
func (pi *PortIter) Reset() error {
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
		p := &Port{Prot: Protocol(hs), Port: uint(port)}
		tmp[cursor] = p
		cursor += 1
	}
	pg.ports = tmp[:cursor]
	return nil
}

// Iter 返回Port迭代器
func (pg *PortGetter) Iter() (*PortIter, error) {
	if pg.ports == nil || len(pg.ports) < 1 {
		return nil, errors.New("没有需要扫描的端口号")
	}
	return &PortIter{ports: pg.ports}, nil
}

// NewPortGetter 返回已初始化的PortGetter
func NewPortGetter() (pg *PortGetter) {
	pg = &PortGetter{filename: portFile}
	pg.Init()
	return
}
