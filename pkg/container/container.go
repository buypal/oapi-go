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

// Container is a wrapper around gabs.Container.
// Overall you probably dont need this wrapper. Use gabs directly.
type Container struct {
	c    *gabs.Container
	path string
}

// Zero returns empty container.
func Zero() Container {
	return Container{}
}

// New returns new container, basically empty map wrapped in gabs.
func New() Container {
	return Container{
		c: gabs.New(),
	}
}

// Make creates new container from interface, if given argument is
// already container it will clone its content and return container
// directly
func Make(v interface{}) (Container, error) {
	x, err := clone(v)
	if err != nil {
		return Container{}, err
	}
	return wrap(x), nil
}

// ReadJSON will read data and return new container.
func ReadJSON(data []byte) (c Container, err error) {
	var i map[string]interface{}
	err = json.Unmarshal(data, &i)
	if err != nil {
		err = errors.Wrapf(err, "failed to unmarshal json")
		return
	}
	c = wrap(i)
	return
}

// ReadYAML will read yaml and returns its content wrapped in container.
func ReadYAML(data []byte) (c Container, err error) {
	var i map[string]interface{}
	err = yaml.Unmarshal(data, &i)
	if err != nil {
		err = errors.Wrapf(err, "failed to unmarshal yaml")
		return
	}
	c = wrap(mmap(i))
	return
}

// ReadFile reads extension and unmarshal .yaml, .yml, .json to
// container. Also stores path of given file, later available with FilePath().
func ReadFile(file string) (c Container, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return Container{}, errors.Wrapf(err, "failed to read file")
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

// Flatten a JSON array or object into an object of key/value pairs for each
// field, where the key is the full path of the structured field in dot path
// notation matching the spec for the method Path.
//
// E.g. the structure `{"foo":[{"bar":"1"},{"bar":"2"}]}` would flatten into the
// object: `{"foo.0.bar":"1","foo.1.bar":"2"}`. `{"foo": [{"bar":[]},{"bar":{}}]}`
// would flatten into the object `{}`
//
// Returns an error if the target is not a JSON object or array.
func (c Container) Flatten() (map[string]interface{}, error) {
	return c.c.Flatten()
}

// ChildrenMap returns a map of all the children of an object element. IF the
// underlying value isn't a object then an empty map is returned.
func (c Container) ChildrenMap() (map[string]Container, error) {
	data := c.Data()
	x := make(map[string]Container)
	switch data.(type) {
	case map[string]interface{}:
	case nil:
		return x, nil
	default:
		return nil, errors.Errorf("children map cannot be called on non-map type")
	}
	for k, v := range c.c.ChildrenMap() {
		x[k] = Container{c: v, path: c.path}
	}
	return x, nil
}

// ExistsP checks whether a dot notation path exists.
func (c Container) ExistsP(path string) bool {
	return c.c.ExistsP(path)
}

// Data returns the underlying value of the target element in the wrapped
// structure.
func (c Container) Data() interface{} {
	return c.c.Data()
}

// IsNil reports if underlying value of container is nil.
func (c Container) IsNil() bool {
	return c.Data() == nil || reflect.ValueOf(c.Data()).IsNil()
}

// FilePath returns file path of given container if was ReadFile was used.
func (c Container) FilePath() string {
	return c.path
}

// Path searches the wrapped structure following a path in dot notation,
// segments of this path are searched according to the same rules as Search.
//
// Because the characters '~' (%x7E) and '.' (%x2E) have special meanings in
// this pkg paths, '~' needs to be encoded as '~0' and '.' needs to be encoded as
// '~1' when these characters appear in a reference key.
func (c Container) Path(x string) Container {
	return Container{c: c.c.Path(x), path: c.path}
}

// SetP sets the value of a field at a path using dot notation, any parts
// of the path that do not exist will be constructed, and if a collision occurs
// with a non object type whilst iterating the path an error is returned.
// Unline original method of gabs it will clone the source value, making
// sure one container dosnt affect other one.
func (c Container) SetP(path string, value interface{}) (err error) {
	val, err := clone(value)
	if err != nil {
		return
	}
	_, err = c.c.SetP(val, path)
	return
}

// Bytes marshals an element to a JSON []byte blob.
func (c Container) Bytes() []byte {
	return c.c.Bytes()
}

// MarshalJSON will marshal given data into json.
func (c Container) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data())
}

// MarshalIndentJSON same oas what would you expect from stdlib.
func (c Container) MarshalIndentJSON(prefix string, indent string) ([]byte, error) {
	return json.MarshalIndent(c.c.Data(), prefix, indent)
}

// MarshalYAML  uses  gopkg.in/yaml.v2 to povide yaml serialization.
func (c Container) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(c.c.Data())
}

// Clone will clone container. Returning its copy.
// This is done using Marshal & Unmarhsal of given container.
func (c Container) Clone() Container {
	cx, _ := clone(c.c.Data())
	return wrap(cx)
}

// Print will print yaml definition of given container to Printer.
func (c Container) Print(log logging.Printer) {
	data, err := yaml.Marshal(c.c.Data())
	if err != nil {
		logging.Debug(log, err.Error())
	} else {
		logging.Debug(log, "%s", data)
	}
}

// Containers list of containers, allowing to do some additional magic.
type Containers []Container

var defaultSearchPatterns = []string{"*.yaml", "*.yml", "*.json"}

// ReadDir will read director and parse yaml/json files.
// It is using glob to search for file in given directory.
// by defaut it will search for yaml,yml,json extension.
// But you are free to supply your own if you have different extension.
func ReadDir(dir string, patterns ...string) (cc Containers, err error) {
	var ff []string
	if len(patterns) == 0 {
		patterns = defaultSearchPatterns
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

// sort interface complience with sort.Sort method

func (cc Containers) Len() int           { return len(cc) }
func (cc Containers) Swap(i, j int)      { cc[i], cc[j] = cc[j], cc[i] }
func (cc Containers) Less(i, j int) bool { return cc[i].path < cc[j].path }

// Sort will clone containers and return is sorted copy
func (cc Containers) Sort() Containers {
	a := make(Containers, len(cc))
	copy(a, cc)
	sort.Sort(a)
	return a
}

// Merge will allow you to merge multiple containers together.
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
