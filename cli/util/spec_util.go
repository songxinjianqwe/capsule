package util

import (
	"encoding/json"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"os"
	"path/filepath"
)

func LoadSpec(bundle string) (spec *specs.Spec, err error) {
	// 如果bundle不为空，则open路径为bundle下的config.json
	// 如果为空，那么open默认是在当前路径下打开
	path := constant.ContainerConfigFilename
	if bundle != "" {
		path = filepath.Join(bundle, path)
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file %s not found, bundle is %s", constant.ContainerConfigFilename, bundle)
		}
		return nil, err
	}
	defer file.Close()
	if err = json.NewDecoder(file).Decode(&spec); err != nil {
		return nil, err
	}
	return spec, validateProcessSpec(spec.Process)
}

func validateProcessSpec(spec *specs.Process) error {
	if spec.Cwd == "" {
		return fmt.Errorf("cwd property must not be empty")
	}
	if !filepath.IsAbs(spec.Cwd) {
		return fmt.Errorf("cwd must be an absolute path")
	}
	if len(spec.Args) == 0 {
		return fmt.Errorf("args must not be empty")
	}
	return nil
}
