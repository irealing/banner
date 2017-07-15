package config

import (
	"errors"
	"github.com/qiniu/log"
)

const (
	defaultConcurrent uint   = 10
	emptyString       string = ""
	defaultPortFile   string = "ports.csv"
	errorInfo         string = "参数异常"
)

var (
	DefaultConfig = &AppConfig{}
)

func init() {
	ap := ArgsParser{args: DefaultConfig}
	ap.Init()
	if err := ap.Parse(); err != nil {
		log.Fatal(err)
	}
}

type AppConfig struct {
	Go     uint `param:"go" usage:"并发数"`
	Input  string `param:"if" usage:"URL地址列表文件"`
	Output string `param:"of" usage:"输出文件"`
	Log    string `param:"log" usage:"日志级别"`
	Port   string `param:"port" usage:"端口列表文件"`
}

func (ac *AppConfig) Validate() (err error) {
	if ac.Go == 0 {
		ac.Go = defaultConcurrent
	}
	if ac.Port == emptyString {
		ac.Port = defaultPortFile
	}
	if ac.Input == emptyString || ac.Output == emptyString {
		err = errors.New(errorInfo)
	}
	logLevel := map[string]int{"debug": log.Ldebug, "info": log.Linfo, "warn": log.Lwarn}
	level, ok := logLevel[ac.Log]
	if !ok {
		level = log.Linfo
	}
	log.SetOutputLevel(level)
	return
}
