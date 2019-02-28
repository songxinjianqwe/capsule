package util

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CleanPath makes a path safe for use with filepath.Join. This is done by not
// only cleaning the path, but also (if the path is relative) adding a leading
// '/' and cleaning it (then removing the leading '/'). This ensures that a
// path resulting from prepending another path will always resolve to lexically
// be a subdirectory of the prefixed path. This is all done lexically, so paths
// that include symlinks won't be safe as a result of using CleanPath.
func CleanPath(path string) string {
	// Deal with empty strings nicely.
	if path == "" {
		return ""
	}

	// Ensure that all paths are cleaned (especially problematic ones like
	// "/../../../../../" which can cause lots of issues).
	path = filepath.Clean(path)

	// If the path isn't absolute, we need to do more processing to fix paths
	// such as "../../../../<etc>/some/path". We also shouldn't convert absolute
	// paths to relative ones.
	if !filepath.IsAbs(path) {
		path = filepath.Clean(string(os.PathSeparator) + path)
		// This can't fail, as (by definition) all paths are relative to root.
		path, _ = filepath.Rel(string(os.PathSeparator), path)
	}

	// Clean the path again for good measure.
	return filepath.Clean(path)
}

// GetAnnotations returns the bundle path and user defined annotations from the
// libcapsule state.  We need to remove the bundle because that is a label
// added by libcapsule.
func GetAnnotations(labels []string) (bundle string, userAnnotations map[string]string) {
	userAnnotations = make(map[string]string)
	for _, l := range labels {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) < 2 {
			continue
		}
		if parts[0] == "bundle" {
			bundle = parts[1]
		} else {
			userAnnotations[parts[0]] = parts[1]
		}
	}
	return
}

func PrintSubsystemPids(subsystemName, cgroupName, context string, init bool) {
	bytes, err := ioutil.ReadFile(path.Join("/sys/fs/cgroup", subsystemName, cgroupName, "tasks"))
	if err != nil {
		logrus.Errorf("read pids failed, cause: %s", err.Error())
	}
	if init {
		logrus.WithField("init", true).Infof("[Pids of %s in %s]%s, context is %s", cgroupName, subsystemName, string(bytes), context)
	} else {
		logrus.Infof("[Pids of %s in %s]%s, context is %s", cgroupName, subsystemName, string(bytes), context)
	}
}
