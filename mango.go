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
	operator     string
	field        string
	value        interface{}
	subselectors []*Selector
}

type combinationNode struct {
	operator string
	nodes    []*interface{}
}

type comparisonNode struct {
	operator string
	field    string
	value    interface{}
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
	if len(doc) > 1 {
		subselectors, err := createSubselectorArray(doc)
		return &Selector{operator: opAnd, subselectors: subselectors}, err
	}
	for key, val := range doc {
		return conditionSelector(key, val)
	}
	return &Selector{}, nil
}

func isObject(i interface{}) bool {
	_, ok := i.(map[string]interface{})
	return ok
}

func createSubselectorArray(i map[string]interface{}) ([]*Selector, error) {
	subselectors := make([]*Selector, 0, len(i))
	for key, value := range i {
		subselector, err := createSubselector(key, value)
		if err != nil {
			return nil, err
		}
		subselectors = append(subselectors, subselector)
	}
	return subselectors, nil
}

func createSubselector(key string, value interface{}) (*Selector, error) {
	if isCombinationOperator(key) {
		return combinationSelector(key, value)
	}
	return conditionSelector(key, value)
}

func combinationSelector(operator string, i interface{}) (*Selector, error) {
	array, ok := i.([]interface{})
	if !ok {
		return nil, errors.Errorf("bad argument for operator %s: <<%v>>", operator, i)
	}
	selectorMap := make(map[string]interface{}, len(array))
	for _, obj := range array {
		if objMap, ok := obj.(map[string]interface{}); ok {
			if len(objMap) > 1 {
				return nil, errors.New("Unimplemented")
			}
			for key, value := range objMap {
				selectorMap[key] = value
			}
		} else {
			return nil, errors.New("unimplemented")
		}
	}
	subselectors, err := createSubselectorArray(selectorMap)
	if err != nil {
		return nil, err
	}
	return &Selector{operator: operator, subselectors: subselectors}, nil
}

type tuple struct {
	key   string
	value interface{}
}

func tupleArray(i interface{}) []tuple {
	switch t := i.(type) {
	case map[string]interface{}:
		tuples := make([]tuple, 0, len(t))
		for key, value := range t {
			tuples = append(tuples, tuple{key: key, value: value})
		}
		return tuples
	case []interface{}:
		tuples := make([]tuple, 0, len(t))
		for _, obj := range t {
			if objMap, ok := obj.(map[string]interface{}); ok {
				if len(objMap) == 1 {
					for key, value := range objMap {
						tuples = append(tuples, tuple{key: key, value: value})
					}
				} else {
					// unimplemented
				}
			} else {
				// unimplemented
			}
		}
		return tuples
	}
	// should never happen
	return nil
}

func conditionSelector(field string, i interface{}) (*Selector, error) {
	if isObject(i) { // explicit
		sel, err := explicitConditionSelector(field, i)
		if err != nil {
			return nil, err
		}
		return sel, nil
	}
	// implicit equality
	return &Selector{operator: opEq, field: field, value: i}, nil
}

func explicitConditionSelector(field string, i interface{}) (*Selector, error) {
	obj, _ := i.(map[string]interface{})
	for k, v := range obj {
		if isConditionOperator(k) {
			return &Selector{operator: k, field: field, value: v}, nil
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

//
// // UnmarshalJSONx unmarshals a JSON selector as described in the CouchDB
// // documentation.
// // http://docs.couchdb.org/en/2.0.0/api/database/find.html#selector-syntax
// func (s *Selector) UnmarshalJSONx(data []byte) error {
// 	var x map[string]json.RawMessage
// 	if err := json.Unmarshal(data, &x); err != nil {
// 		return err
// 	}
// 	if len(x) == 0 {
// 		return nil
// 	}
// 	var sels []Selector
// 	for k, v := range x {
// 		var op string
// 		var field string
// 		var value interface{}
// 		field = k
// 		if v[0] == '{' {
// 			var e error
// 			op, value, e = opPattern(v)
// 			if e != nil {
// 				return e
// 			}
// 		}
// 		if op == "" {
// 			op = opEq
// 			if e := json.Unmarshal(v, &value); e != nil {
// 				return e
// 			}
// 		}
// 		sels = append(sels, Selector{
// 			op:    op,
// 			field: field,
// 			value: value,
// 		})
// 	}
// 	if len(sels) == 1 {
// 		*s = sels[0]
// 	} else {
// 		*s = Selector{
// 			op:  opAnd,
// 			sel: sels,
// 		}
// 	}
// 	return nil
// }

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
	switch s.operator {
	case opNone:
		return true, nil
	case opEq, opGT, opGTE, opLT, opLTE:
		v, ok := doc[s.field]
		if !ok {
			return false, nil
		}
		switch s.operator {
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
		for _, sel := range s.subselectors {
			m, e := sel.Matches(doc)
			if e != nil || !m {
				return m, e
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown mango operator '%s'", s.operator)
	}
	return true, nil
}
