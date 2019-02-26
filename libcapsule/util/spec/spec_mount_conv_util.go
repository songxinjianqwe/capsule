package spec

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"golang.org/x/sys/unix"
	"path/filepath"
	"strings"
)

var flags = map[string]struct {
	clear bool
	flag  int
}{
	"acl":           {false, unix.MS_POSIXACL},
	"async":         {true, unix.MS_SYNCHRONOUS},
	"atime":         {true, unix.MS_NOATIME},
	"bind":          {false, unix.MS_BIND},
	"defaults":      {false, 0},
	"dev":           {true, unix.MS_NODEV},
	"diratime":      {true, unix.MS_NODIRATIME},
	"dirsync":       {false, unix.MS_DIRSYNC},
	"exec":          {true, unix.MS_NOEXEC},
	"iversion":      {false, unix.MS_I_VERSION},
	"lazytime":      {false, unix.MS_LAZYTIME},
	"loud":          {true, unix.MS_SILENT},
	"mand":          {false, unix.MS_MANDLOCK},
	"noacl":         {true, unix.MS_POSIXACL},
	"noatime":       {false, unix.MS_NOATIME},
	"nodev":         {false, unix.MS_NODEV},
	"nodiratime":    {false, unix.MS_NODIRATIME},
	"noexec":        {false, unix.MS_NOEXEC},
	"noiversion":    {true, unix.MS_I_VERSION},
	"nolazytime":    {true, unix.MS_LAZYTIME},
	"nomand":        {true, unix.MS_MANDLOCK},
	"norelatime":    {true, unix.MS_RELATIME},
	"nostrictatime": {true, unix.MS_STRICTATIME},
	"nosuid":        {false, unix.MS_NOSUID},
	"rbind":         {false, unix.MS_BIND | unix.MS_REC},
	"relatime":      {false, unix.MS_RELATIME},
	"remount":       {false, unix.MS_REMOUNT},
	"ro":            {false, unix.MS_RDONLY},
	"rw":            {true, unix.MS_RDONLY},
	"silent":        {false, unix.MS_SILENT},
	"strictatime":   {false, unix.MS_STRICTATIME},
	"suid":          {true, unix.MS_NOSUID},
	"sync":          {false, unix.MS_SYNCHRONOUS},
}

func createMount(cwd string, specMount specs.Mount) *configc.Mount {
	logrus.Infof("converting specs.mount to configc.Mount...")
	flags, data := parseMountOptions(specMount.Options)
	source := specMount.Source
	device := specMount.Type
	if flags&unix.MS_BIND != 0 {
		if device == "" {
			device = "bind"
		}
		if !filepath.IsAbs(source) {
			source = filepath.Join(cwd, specMount.Source)
		}
	}
	return &configc.Mount{
		// device是type，比如proc、tmpfs
		Device:      device,
		Source:      source,
		Destination: specMount.Destination,
		Data:        data,
		Flags:       flags,
	}
}

// parseMountOptions parses the string and returns the flags and any mount data that it contains.
func parseMountOptions(options []string) (int, string) {
	var (
		flag int
		data []string
	)
	for _, o := range options {
		if f, exists := flags[o]; exists && f.flag != 0 {
			if f.clear {
				flag &= ^f.flag
			} else {
				flag |= f.flag
			}
		} else {
			data = append(data, o)
		}
	}
	return flag, strings.Join(data, ",")
}
