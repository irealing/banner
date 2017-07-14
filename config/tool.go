package config

import (
	"flag"
	"errors"
	"reflect"
)

const (
	paramKey = "param"
)

var (
	notPointer = errors.New("不支持的类型，不是指针")
)

type Arguments interface {
	Validate() error
}
type ArgsParser struct {
	args   Arguments
	values []interface{}
}

func (ap *ArgsParser) Init() error {
	return ap.register()
}
func (ap *ArgsParser) register() error {
	argv := reflect.ValueOf(ap.args)
	if argv.Type().Kind() != reflect.Ptr {
		return notPointer
	}
	argv = argv.Elem()
	at := argv.Type()
	var err error
	fields := at.NumField()
	ap.values = make([]interface{}, fields)
	for i := 0; i < fields; i++ {
		ft := at.Field(i)
		//fv := argv.Field(i)
		ap.regField(i, &ft)
	}
	return err
}
func (ap *ArgsParser) regField(i int, st *reflect.StructField) {
	param := st.Tag.Get(paramKey)
	if param == emptyString {
		return
	}
	switch st.Type.Kind() {
	case reflect.String:
		ap.values[i] = flag.String(param, emptyString, emptyString)
	case reflect.Int:
		ap.values[i] = flag.Int(param, 0, emptyString)
	case reflect.Uint:
		ap.values[i] = flag.Uint(param, 0, emptyString)
	}
}
func (ap *ArgsParser) rejectValues() {
	value := reflect.ValueOf(ap.args).Elem()
	for i, v := range ap.values {
		switch v.(type) {
		case nil:
			continue
		case *string:
			s := v.(*string)
			value.Field(i).SetString(*s)
		case *int:
			iv := v.(*int)
			value.Field(i).SetInt(int64(*iv))
		case *uint:
			uiv := v.(*uint)
			value.Field(i).SetUint(uint64(*uiv))
		}
	}
}

func (ap *ArgsParser) Parse() error {
	flag.Parse()
	ap.rejectValues()
	err := ap.args.Validate()
	if err != nil {
		flag.PrintDefaults()
	}
	return err
}
