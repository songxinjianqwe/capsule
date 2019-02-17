package validate

import (
	"github.com/songxinjianqwe/rune/libcapsule/config"
)

type Validator interface {
	Validate(*config.Config) error
}

func New() Validator {
	return &ConfigValidator{}
}

type ConfigValidator struct {
}

func (v *ConfigValidator) Validate(config *config.Config) error {
	return nil
}
