package config

import "flag"

type Arguments interface {
	Validate() error
}
type ArgsParser struct {
	args Arguments
}

func (ap *ArgsParser) Init() error {
	return ap.register()
}
func (ap *ArgsParser) register() error {
	return nil
}
func (ap *ArgsParser) Parse() error {
	err := ap.args.Validate()
	if err != nil {
		flag.PrintDefaults()
	}
	return err
}
