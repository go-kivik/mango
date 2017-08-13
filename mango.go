package mango

import (
	"encoding/json"
	"fmt"

	"github.com/flimzy/kivik/collate"
	"github.com/pkg/errors"
)

// Selector represents a CouchDB Find query selector. See
// http://docs.couchdb.org/en/2.0.0/api/database/find.html#find-selectors
type Selector struct {
	op    string
	field string
	value interface{}
	sel   []Selector
}

// New returns a new selector, parsed from data.
func New(data string) (*Selector, error) {
	s := &Selector{}
	err := json.Unmarshal([]byte(data), &s)
	return s, err
}

// UnmarshalJSON xxx ...
func (s *Selector) UnmarshalJSON(data []byte) error {
	doc, err := parseSelectorInput(data)
	if err != nil {
		return err
	}

	sel, err := createSelector(doc)
	if err != nil {
		return err
	}
	*s = *sel
	return nil
}

func createSelector(doc map[string]interface{}) (*Selector, error) {
	var op, field string
	var value interface{}
	for key, val := range doc {
		if isObject(val) { // explicit
			sel, err := explicitConditionSelector(key, val)
			if err != nil {
				return nil, err
			}
			return sel, nil
		}
		// implicit equality
		op = opEq
		field = key
		value = val
	}
	return &Selector{op: op, field: field, value: value}, nil
}

func isObject(i interface{}) bool {
	_, ok := i.(map[string]interface{})
	return ok
}

func explicitConditionSelector(field string, i interface{}) (*Selector, error) {
	obj, _ := i.(map[string]interface{})
	for k, v := range obj {
		if isConditionOperator(k) {
			return &Selector{op: k, field: field, value: v}, nil
		}
	}
	return nil, errors.New("subfields not implemented")
}

func parseSelectorInput(data []byte) (map[string]interface{}, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc, validateKeys(doc)
}

func populateSelector(selector *Selector, doc map[string]interface{}) error {
	return nil
}

func validateKeys(doc map[string]interface{}) error {
	for key, value := range doc {
		if isOperator(key) && !isSupportedOperator(key) {
			return errors.Errorf("unknown mango operator '%s'", key)
		}

		switch t := value.(type) {
		case map[string]interface{}:
			if err := validateKeys(t); err != nil {
				return err
			}
		case []interface{}:
			for _, val := range t {
				if obj, ok := val.(map[string]interface{}); ok {
					if err := validateKeys(obj); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// UnmarshalJSONx unmarshals a JSON selector as described in the CouchDB
// documentation.
// http://docs.couchdb.org/en/2.0.0/api/database/find.html#selector-syntax
func (s *Selector) UnmarshalJSONx(data []byte) error {
	var x map[string]json.RawMessage
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if len(x) == 0 {
		return nil
	}
	var sels []Selector
	for k, v := range x {
		var op string
		var field string
		var value interface{}
		field = k
		if v[0] == '{' {
			var e error
			op, value, e = opPattern(v)
			if e != nil {
				return e
			}
		}
		if op == "" {
			op = opEq
			if e := json.Unmarshal(v, &value); e != nil {
				return e
			}
		}
		sels = append(sels, Selector{
			op:    op,
			field: field,
			value: value,
		})
	}
	if len(sels) == 1 {
		*s = sels[0]
	} else {
		*s = Selector{
			op:  opAnd,
			sel: sels,
		}
	}
	return nil
}

func opPattern(data []byte) (op string, value interface{}, err error) {
	var x map[string]json.RawMessage
	if e := json.Unmarshal(data, &x); e != nil {
		return "", nil, e
	}
	if len(x) != 1 {
		panic("got more than one result")
	}
	for k, v := range x {
		switch k {
		case opEq, opNE, opLT, opLTE, opGT, opGTE:
			var value interface{}
			if e := json.Unmarshal(v, &value); e != nil {
				return "", nil, e
			}
			return k, value, nil
		default:
			return "", nil, fmt.Errorf("unknown mango operator '%s'", k)
		}
	}
	return "", nil, nil
}

type couchDoc map[string]interface{}

// Matches returns true if the provided doc matches the selector.
func (s *Selector) Matches(doc couchDoc) (bool, error) {
	c := &collate.Raw{}
	switch s.op {
	case opNone:
		return true, nil
	case opEq, opGT, opGTE, opLT, opLTE:
		v, ok := doc[s.field]
		if !ok {
			return false, nil
		}
		switch s.op {
		case opEq:
			return c.Eq(v, s.value), nil
		case opGT:
			return c.GT(v, s.value), nil
		case opGTE:
			return c.GTE(v, s.value), nil
		case opLT:
			return c.LT(v, s.value), nil
		case opLTE:
			return c.LTE(v, s.value), nil
		}
	case opAnd:
		for _, sel := range s.sel {
			m, e := sel.Matches(doc)
			if e != nil || !m {
				return m, e
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown mango operator '%s'", s.op)
	}
	return true, nil
}
