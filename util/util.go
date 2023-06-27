package util

import (
	"context"
	"github.com/pkg/errors"
	"reflect"
	"regexp"
)

// todo: be common with official module
func GetCaptureGroupMap(re *regexp.Regexp, ma []string) map[string]string {
	cgm := make(map[string]string)
	for idx, name := range re.SubexpNames() {
		if idx == 0 {
			continue
		}
		cgm[name] = ma[idx]
	}

	return cgm
}

// todo: be common with official module
func Struct2HashFields(_ context.Context, s any) (map[string]any, error) {
	v := reflect.Indirect(reflect.ValueOf(s))
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return nil, errors.Errorf("not a struct")
	}

	m := make(map[string]any)
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		vf := v.Field(i)
		if !vf.CanInterface() {
			return nil, errors.Errorf("field:%s cannot got it interface", sf.Name)
		}
		m[sf.Name] = vf.Interface()
	}

	return m, nil
}
