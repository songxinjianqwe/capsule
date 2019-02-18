package validate

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
)

type Validator interface {
	Validate(*configc.Config) error
}

func New() Validator {
	return &ConfigValidator{}
}

type ConfigValidator struct {
}

func (v *ConfigValidator) Validate(config *configc.Config) error {
	return nil
}
