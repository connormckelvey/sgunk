package tree

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/connormckelvey/ssg/util"
)

type NodeAttributes struct {
	attr map[any][]map[string]any
	err  error
}

func newNodeAttributes() *NodeAttributes {
	return &NodeAttributes{
		attr: make(map[any][]map[string]any),
	}
}

func (a *NodeAttributes) Add(namespace any, attributes any) error {
	m, err := util.MarshalMap(attributes)
	if err != nil {
		return err
	}
	a.attr[namespace] = append(a.attr[namespace], m)
	return nil
}

func (a *NodeAttributes) Get(namespace any) (map[string]any, bool) {
	values, ok := a.attr[namespace]
	if !ok {
		return nil, false
	}

	result := make(map[string]any)
	for _, v := range values {
		err := mergo.Merge(&result, v)
		if err != nil {
			a.err = err
			return nil, false
		}
	}
	return result, true
}

func (a *NodeAttributes) GetAll(namespace any) ([]map[string]any, bool) {
	values, ok := a.attr[namespace]
	if !ok {
		return nil, false
	}
	return values, true
}

func (a *NodeAttributes) Map() (map[string]map[string]any, error) {
	m := make(map[string]map[string]any)
	for ns := range a.attr {
		key := fmt.Sprint(ns)
		v, ok := a.Get(ns)
		if !ok {
			return nil, a.err
		}
		m[key] = v
	}
	return m, nil
}
