package cache

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrTargetNotPointer = errors.New("target must be a non-nil pointer")

func assignReflect(val any, target any) error {
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Pointer || targetVal.IsNil() {
		return ErrTargetNotPointer
	}

	elem := targetVal.Elem()
	if !elem.CanSet() {
		return errors.New("target element cannot be set")
	}

	if val == nil {
		elem.Set(reflect.Zero(elem.Type()))
		return nil
	}

	srcVal := reflect.ValueOf(val)
	if srcVal.Type().AssignableTo(elem.Type()) {
		elem.Set(srcVal)
		return nil
	}

	if srcVal.Type().ConvertibleTo(elem.Type()) {
		elem.Set(srcVal.Convert(elem.Type()))
		return nil
	}

	return fmt.Errorf("cannot assign type %s to target type %s", srcVal.Type(), elem.Type())
}
