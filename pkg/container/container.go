package container

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/Jeffail/gabs/v2"
	"github.com/buypal/oapi-go/pkg/logging"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Container struct {
	c    *gabs.Container
	path string
}

func Zero() Container {
	return Container{}
}

func EmptyObject() Container {
	return Wrap(make(map[string]interface{}))
}

func Wrap(c interface{}) Container {
	return Container{
		c: gabs.Wrap(c),
	}
}

func New() Container {
	return Container{
		c: gabs.New(),
	}
}

func Make(v interface{}) (Container, error) {
	x, err := clone(v)
	if err != nil {
		return Container{}, nil
	}
	return Wrap(x), nil
}

func ReadJSON(data []byte) (c Container, err error) {
	var i map[string]interface{}
	err = json.Unmarshal(data, &i)
	if err != nil {
		return
	}
	c = Wrap(i)
	return
}

func ReadYAML(data []byte) (c Container, err error) {
	var i map[string]interface{}
	err = yaml.Unmarshal(data, &i)
	if err != nil {
		return
	}
	c = Wrap(mmap(i))
	return
}

func ReadFile(file string) (c Container, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return Container{}, err
	}
	ext := filepath.Ext(file)
	switch ext {
	case ".yaml", ".yml":
		c, err = ReadYAML(data)
	case ".json":
		c, err = ReadJSON(data)
	default:
		err = errors.Errorf("failed to recognize extension %q for unmarshal", ext)
	}
	if err != nil {
		return
	}
	c.path = file
	return
}

func (c Container) Flatten() (map[string]interface{}, error) {
	return c.c.Flatten()
}

func (c Container) ChildrenMap() (map[string]Container, error) {
	data := c.c.Data()
	x := make(map[string]Container)
	switch data.(type) {
	case map[string]interface{}:
	case nil:
		return x, nil
	default:
		return nil, errors.Errorf("children map cannot be called on non-map type")
	}
	for k, v := range c.c.ChildrenMap() {
		x[k] = Container{c: v}
	}
	return x, nil
}

func (c Container) ExistsP(path string) bool {
	return c.c.ExistsP(path)
}

func (c Container) Data() interface{} {
	return c.c.Data()
}

func (c Container) IsNil() bool {
	return c.Data() == nil || reflect.ValueOf(c.Data()).IsNil()
}

func (c Container) FilePath() string {
	return c.path
}

func (c Container) Path(x string) Container {
	return Container{c: c.c.Path(x), path: c.path}
}

func (c Container) PathF(x string, fallback Container) Container {
	p := c.Path(x)
	if p.IsNil() {
		return fallback
	}
	return p
}

func (c Container) SetP(path string, value interface{}) (cs Container, err error) {
	val, err := clone(value)
	if err != nil {
		return
	}
	cx, err := c.c.SetP(val, path)
	if err != nil {
		return
	}
	cs = Container{c: cx}
	return
}

func (c Container) Bytes() []byte {
	return c.c.Bytes()
}

func (c Container) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c.Data())
}

func (c Container) MarshalIndentJSON(prefix string, indent string) ([]byte, error) {
	return json.MarshalIndent(c.c.Data(), prefix, indent)
}

func (c Container) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(c.c.Data())
}

func (c Container) Clone() Container {
	cx, _ := clone(c.c.Data())
	return Wrap(cx)
}

func (c Container) Print(log logging.Printer) {
	data, err := yaml.Marshal(c.c.Data())
	if err != nil {
		logging.LogFunc(log)(err.Error())
	} else {
		logging.LogFunc(log)("%s", data)
	}
}

type Containers []Container

func ReadDir(dir string, patterns ...string) (cc Containers, err error) {
	var ff []string
	if len(patterns) == 0 {
		patterns = []string{"*.yaml", "*.yml", "*.json"}
	}

	for _, s := range patterns {
		xs, err := filepath.Glob(filepath.Join(dir, s))
		if err != nil {
			return nil, err
		}
		ff = append(ff, xs...)
	}

	if len(ff) == 0 {
		return
	}

	for _, m := range ff {
		c, err := ReadFile(m)
		if err != nil {
			return cc, err
		}
		cc = append(cc, c)
	}

	return
}

// Sort interface complience
func (cc Containers) Len() int           { return len(cc) }
func (cc Containers) Swap(i, j int)      { cc[i], cc[j] = cc[j], cc[i] }
func (cc Containers) Less(i, j int) bool { return cc[i].path < cc[j].path }

func (cc Containers) Sort() Containers {
	a := make(Containers, len(cc))
	copy(a, cc)
	sort.Sort(a)
	return a
}

func (cc Containers) Merge(fn Merger) (Container, error) {
	c := New()
	for _, x := range cc {
		err := c.Merge(x, fn)
		if err != nil {
			return Container{}, err
		}
	}
	return c, nil
}
