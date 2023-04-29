package dockertool

import (
	"reflect"
	"regexp"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/KDGoLib/errutil"
)

// errors
var (
	ErrUnsupportedContainerType1 = errutil.NewFactory("unsupported container type: %q")
)

var (
	regNameTrim = regexp.MustCompile(`^/`)
)

// GetContainerInfo return container info from docker object
func GetContainerInfo(container any) (id string, name string, err error) {
	switch info := container.(type) {
	case docker.APIContainers:
		return info.ID, regNameTrim.ReplaceAllString(info.Names[0], ""), nil
	case *docker.Container:
		return info.ID, regNameTrim.ReplaceAllString(info.Name, ""), nil
	default:
		return "", "", ErrUnsupportedContainerType1.New(nil, reflect.TypeOf(container).String())
	}
}
