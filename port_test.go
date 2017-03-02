package main

import (
	"testing"
)

const (
	csvName = "ports.csv"
)

// TestPortGetterInit 测试PortGetter初始化函数
func TestPortGetterInit(t *testing.T) {
	pg := PortGetter{filename: csvName}
	err := pg.Init()
	if err == nil {
		t.Log("TestPortGetterInit success")
	} else {
		t.Error("TestPortGetterInit failed ")
	}
	if pg.Size() != 31 {
		t.Fatalf("pg size is %d", pg.Size())
	}
}

// TestPortIter 测试Port迭代器
func TestPortIter(t *testing.T) {
	pg := PortGetter{filename: csvName}
	err := pg.Init()
	if err != nil {
		t.Fatal(err)
	}
	iter, err := pg.Iter()
	if err != nil {
		t.Fatal(err)
	}
	for iter.HasNext() {
		p := iter.Next()
		t.Logf("protocol:%s\t port: %d \n", p.Prot, p.Port)
	}
}
