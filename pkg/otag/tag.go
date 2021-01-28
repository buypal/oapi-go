package otag

import (
	"strconv"

	"github.com/fatih/structtag"
	"github.com/vmihailenco/tagparser"
)

// Tag represent oapi tag. Only supporting subset of properties
// which might be handy in both validation & schema attributes supply.
// It is not intended to contain whole scheme definition.
type Tag struct {
	Ignore     bool
	Inline     *bool
	OmitEmpty  bool
	Nullable   *bool
	ReadOnly   bool
	WriteOnly  bool
	Deprecated bool
	Required   bool
	Name       string
	Pattern    string
	Format     string
	Type       string
	Min        *float64
	Max        *float64
	EMin       *float64
	EMax       *float64
	MulOf      *float64
	MinLen     *int64
	MaxLen     *int64
	MinItems   *int64
	MaxItems   *int64
	UniqItems  bool
	MinProps   *int64
	MaxProps   *int64
}

// Parse will parse all fileds and tags
func Parse(rawtags string) (meta Tag, err error) {
	if len(rawtags) == 0 {
		return
	}
	tags, err := structtag.Parse(rawtags)
	if err != nil {
		return
	}
	if err = parseJSONTag(tags, &meta); err != nil {
		return
	}
	if err = parseOAPITag(tags, &meta); err != nil {
		return
	}
	return
}

func parseOAPITag(tags *structtag.Tags, meta *Tag) (err error) {
	tag := parseTag(tags, "oapi")
	if tag == nil {
		return nil
	}
	if tag.Name == "-" {
		meta.Ignore = true
		return
	}
	if len(tag.Name) != 0 {
		meta.Name = tag.Name
	}
	errs := []error{
		parseTagBoolPtr(tag, "inline", &meta.Inline),
		parseTagBool(tag, "omitempty", &meta.OmitEmpty),
		parseTagBoolPtr(tag, "nullable", &meta.Nullable),
		parseTagBool(tag, "readonly", &meta.ReadOnly),
		parseTagBool(tag, "writeonly", &meta.WriteOnly),
		parseTagBool(tag, "deprecated", &meta.Deprecated),
		parseTagBool(tag, "unique", &meta.UniqItems),
		parseTagBool(tag, "required", &meta.Required),
		parseTagIntPtr(tag, "maxlen", &meta.MaxLen),
		parseTagIntPtr(tag, "minlen", &meta.MinLen),
		parseTagIntPtr(tag, "maxitems", &meta.MaxItems),
		parseTagIntPtr(tag, "minitems", &meta.MinItems),
		parseTagIntPtr(tag, "maxprops", &meta.MaxProps),
		parseTagIntPtr(tag, "minprops", &meta.MinProps),
		parseTagString(tag, "pattern", &meta.Pattern),
		parseTagString(tag, "format", &meta.Format),
		parseTagString(tag, "type", &meta.Type),
		parseTagFloatPtr(tag, "max", &meta.Max),
		parseTagFloatPtr(tag, "min", &meta.Min),
		parseTagFloatPtr(tag, "emax", &meta.EMax),
		parseTagFloatPtr(tag, "emin", &meta.EMin),
		parseTagFloatPtr(tag, "mulof", &meta.EMin),
	}
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return
}

func parseJSONTag(tags *structtag.Tags, meta *Tag) (err error) {
	tag := parseTag(tags, "json")
	if tag == nil {
		return nil
	}
	if tag.Name == "-" {
		meta.Ignore = true
		return
	}
	if len(tag.Name) != 0 {
		meta.Name = tag.Name
	}
	errs := []error{
		parseTagBoolPtr(tag, "inline", &meta.Inline),
		parseTagBool(tag, "omitempty", &meta.OmitEmpty),
	}
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return
}

func parseTag(tags *structtag.Tags, tag string) (result *tagparser.Tag) {
	oapi, err := tags.Get(tag)
	if err != nil {
		return nil
	}
	result = tagparser.Parse(oapi.Value())
	return result
}

func parseTagFloat(m *tagparser.Tag, tag string, val *float64) error {
	v, ok := m.Options[tag]
	if !ok {
		return nil
	}
	x, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return err
	}
	*val = x
	return nil
}

func parseTagFloatPtr(m *tagparser.Tag, tag string, val **float64) error {
	if _, ok := m.Options[tag]; !ok {
		return nil
	}
	var v float64
	err := parseTagFloat(m, tag, &v)
	if err != nil {
		return err
	}
	*val = &v
	return nil
}

func parseTagInt(m *tagparser.Tag, tag string, val *int64) error {
	v, ok := m.Options[tag]
	if !ok {
		return nil
	}
	x, err := strconv.Atoi(v)
	if err != nil {
		return err
	}
	*val = int64(x)
	return nil
}

func parseTagIntPtr(m *tagparser.Tag, tag string, val **int64) error {
	if _, ok := m.Options[tag]; !ok {
		return nil
	}
	var v int64
	err := parseTagInt(m, tag, &v)
	if err != nil {
		return err
	}
	*val = &v
	return nil
}

func parseTagString(m *tagparser.Tag, tag string, val *string) error {
	v, ok := m.Options[tag]
	if !ok {
		return nil
	}
	*val = v
	return nil
}

func parseTagBool(m *tagparser.Tag, tag string, val *bool) (err error) {
	x, ok := m.Options[tag]
	if x == "" {
		*val = ok
	} else {
		*val, err = strconv.ParseBool(x)
	}
	return nil
}

func parseTagBoolPtr(m *tagparser.Tag, tag string, val **bool) (err error) {
	x, ok := m.Options[tag]
	if !ok {
		return
	}
	if x == "" {
		*val = &ok
	} else {
		var y bool
		y, err = strconv.ParseBool(x)
		if err != nil {
			return
		}
		*val = &y
	}
	return nil
}
