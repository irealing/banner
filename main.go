package main

import (
	"os"

	"github.com/qiniu/log"
)

func main() {
	cfg := DefaultConfig
	start(cfg.Input, cfg.Output, cfg.Go)
}

// work 启动
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
