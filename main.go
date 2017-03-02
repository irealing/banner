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
	_level := map[string]int{"debug": log.Ldebug, "info": log.Linfo, "warn": log.Lwarn, "error": log.Lerror}
	flag.Parse()
	log.SetOutputLevel(_level[*l])
	if *inPath == "" || *outPath == "" || *concurrent < 1 {
		log.Error("输入参数异常:\nif:输入文件;\nof:输出文件;\ngo:启动协程数")
		os.Exit(1)
	}
	start(*inPath, *outPath, uint(*concurrent))
}

// start 启动
func start(i, o string, con uint) {
	// i 输入文件名
	//  o 输出文件名
	//con 并发量
	input, err := os.Open(i)
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	var output *os.File
	if _, err := os.Stat(o); os.IsNotExist(err) {
		output, err = os.Create(o)
	} else {
		output, err = os.OpenFile(o, os.O_RDWR|os.O_APPEND, 0600)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()
	scheduler := NewScheduler(con, input, output)
	count := scheduler.Run()
	log.Info("Exit: ", count)
}
