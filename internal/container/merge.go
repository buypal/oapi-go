package container

import "github.com/pkg/errors"

// Merger is function signature allowing merging
type Merger func(destination, source interface{}) (interface{}, error)

// MergeStrict will merge strictly, no collisions
func MergeStrict(dest, source interface{}) (interface{}, error) {
	return nil, errors.Errorf("%v collied with %v", dest, source)
}

// MergeDefault will merge default
func MergeDefault(dest, source interface{}) (interface{}, error) {
	return dest, nil
}

// MergeOverride will merge with override
func MergeOverride(dest, source interface{}) (interface{}, error) {
	return source, nil
}

// Merge merges two objects using a provided function to resolve collisions.
//
// The collision function receives two interface{} arguments, destination (the
// original object) and source (the object being merged into the destination).
// Which ever value is returned becomes the new value in the destination object
// at the location of the collision.
func (c Container) Merge(source Container, collisionFn Merger) error {
	source = source.Clone() // make sure we are not moving pointers

	var recursiveFnc func(map[string]interface{}, []string) error

	// recursivly merge structures
	recursiveFnc = func(mmap map[string]interface{}, path []string) error {
		for key, value := range mmap {
			newPath := append(path, key)
			if !c.c.Exists(newPath...) {
				// path doesn't exist. So set the value
				if _, err := c.c.Set(value, newPath...); err != nil {
					return err
				}
				continue
			}
			existingData := c.c.Search(newPath...).Data()
			switch t := value.(type) {
			case map[string]interface{}:
				switch existingVal := existingData.(type) {
				case map[string]interface{}:
					err := recursiveFnc(t, newPath)
					if err != nil {
						return err
					}
				default:
					xx, err := collisionFn(existingVal, t)
					if err != nil {
						return errors.Wrapf(err, "at %s", source.path)
					}
					_, err = c.c.Set(xx, newPath...)
					if err != nil {
						return err
					}
				}
			default:
				xx, err := collisionFn(existingData, t)
				if err != nil {
					return errors.Wrapf(err, "at %s", source.path)
				}
				_, err = c.c.Set(xx, newPath...)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	if mmap, ok := source.c.Data().(map[string]interface{}); ok {
		return recursiveFnc(mmap, []string{})
	}
	return nil
}
