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
	logLevel      = map[string]int{"debug": log.Ldebug, "info": log.Linfo, "warn": log.Lwarn}
)

func init() {
	ap := ArgsParser{args: DefaultConfig}
	ap.Init()
	if err := ap.Parse(); err != nil {
		log.Fatal(err)
	}
}

type AppConfig struct {
	Go     uint `param:"go"`
	Input  string `param:"if"`
	Output string `param:"of"`
	Log    string `param:"log"`
	Port   string `param:"port"`
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
	level, ok := logLevel[ac.Log]
	if !ok {
		level = log.Linfo
	}
	log.SetOutputLevel(level)
	return
}
