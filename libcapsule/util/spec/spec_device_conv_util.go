package spec

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configc"
	"os"
)

func stringToDeviceRune(s string) (rune, error) {
	switch s {
	case "p":
		return 'p', nil
	case "u":
		return 'u', nil
	case "b":
		return 'b', nil
	case "c":
		return 'c', nil
	default:
		return 0, fmt.Errorf("invalid device type %q", s)
	}
}

func createDevices(spec *specs.Spec, config *configc.ContainerConfig) error {
	logrus.Infof("creating devices...")
	// add whitelisted devices
	config.Devices = []*configc.Device{
		{
			Type:     'c',
			Path:     "/dev/null",
			Major:    1,
			Minor:    3,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/random",
			Major:    1,
			Minor:    8,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/full",
			Major:    1,
			Minor:    7,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/tty",
			Major:    5,
			Minor:    0,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/zero",
			Major:    1,
			Minor:    5,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/urandom",
			Major:    1,
			Minor:    9,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
	}
	// merge in additional devices from the spec
	if spec.Linux != nil {
		for _, d := range spec.Linux.Devices {
			var uid, gid uint32
			var filemode os.FileMode = 0666

			if d.UID != nil {
				uid = *d.UID
			}
			if d.GID != nil {
				gid = *d.GID
			}
			dt, err := stringToDeviceRune(d.Type)
			if err != nil {
				return err
			}
			if d.FileMode != nil {
				filemode = *d.FileMode
			}
			device := &configc.Device{
				Type:     dt,
				Path:     d.Path,
				Major:    d.Major,
				Minor:    d.Minor,
				FileMode: filemode,
				Uid:      uid,
				Gid:      gid,
			}
			config.Devices = append(config.Devices, device)
		}
	}
	return nil
}
