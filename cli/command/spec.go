package command

import (
	"encoding/json"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/songxinjianqwe/capsule/libcapsule/util/spec"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
)

var SpecCommand = cli.Command{
	Name:  "spec",
	Usage: "create a new specification file",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 0, util.ExactArgs); err != nil {
			return err
		}
		exampleSpec := spec.Example()
		if err := util.CheckNoFile(libcapsule.ContainerConfigFilename); err != nil {
			return err
		}
		data, err := json.MarshalIndent(exampleSpec, "", "\t")
		if err != nil {
			return err
		}
		return ioutil.WriteFile(libcapsule.ContainerConfigFilename, data, 0666)
	},
}

func loadSpec() (spec *specs.Spec, err error) {
	file, err := os.Open(libcapsule.ContainerConfigFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file %s not found", libcapsule.ContainerConfigFilename)
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
