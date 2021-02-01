package spec

import (
	"encoding/json"
	"errors"
)

// Callback It's the Union type of callback and Refable
type Callback struct {
	Refable  `json:",inline"`
	Callback map[Expressions]*PathItem
}

// MarshalJSON returns m as the JSON encoding of callback or Refable.
func (s Callback) MarshalJSON() ([]byte, error) {
	if s.Ref != nil {
		return json.Marshal(s.Refable)
	}
	return json.Marshal(s.Callback)
}

// UnmarshalJSON sets callback or Refable to data.
func (s *Callback) UnmarshalJSON(data []byte) error {
	if s == nil {
		return errors.New("spec.Callback: UnmarshalJSON on nil pointer")
	}
	if len(data) == 0 {
		return nil
	}
	err := json.Unmarshal(data, &s.Refable)
	if err != nil {
		return err
	}
	if s.Ref != nil {
		return nil
	}
	return json.Unmarshal(data, &s.Callback)
}

// Entity satisfies componenter interface
func (s Callback) Entity() Entity {
	return CallbackKind
}
