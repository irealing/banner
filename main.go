package main

import (
	"flag"
	"os"

	"github.com/qiniu/log"
)

func main() {
	inPath := flag.String("if", "", "输入文件")
	outPath := flag.String("of", "", "输出文件")
	concurrent := flag.Int("go", 50, "启动协程数")
	l := flag.String("l", "debug", "日志级别")
	protocol := flag.String("P", "http", "协议")
	_level := map[string]int{"debug": log.Ldebug, "info": log.Linfo, "warn": log.Lwarn, "error": log.Lerror}
	flag.Parse()
	log.SetOutputLevel(_level[*l])
	if *inPath == "" || *outPath == "" || *concurrent < 1 {
		log.Errorf("输入参数异常:\nif:输入文件;of:输出文件;\ngo:启动协程数", *inPath, *outPath, *concurrent)
		os.Exit(1)
	}
	if *protocol != "http" && *protocol != "https" {
		log.Fatalf("错误的协议:%s", *protocol)
		os.Exit(1)
	}
	start(*inPath, *outPath, *protocol, uint(*concurrent))
}

// start 启动
func start(i, o, p string, con uint) {
	// i 输入文件名
	//  o 输出文件名
	// p 协议（http/https）
	//con 并发量
	input, err := os.Open(i)
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	output, err := os.OpenFile(o, os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()
	scheduler := NewScheduler(con, input, output, p)
	count := scheduler.Run()
	log.Info("Exit: ", count)
}
