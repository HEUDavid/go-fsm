package util

import (
	"reflect"
)

func SetAttr(o interface{}, attr string, val interface{}) bool {
	v := reflect.ValueOf(o)
	if v.Kind() != reflect.Ptr {
		return false
	}

	v = v.Elem()

	field := v.FieldByName(attr)
	if !(field.IsValid() && field.CanSet()) {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(val.(string))
		return true
	case reflect.Int:
		field.SetInt(int64(val.(int)))
		return true
	case reflect.Int32:
		field.SetInt(int64(val.(int32)))
		return true
	case reflect.Int64:
		field.SetInt(val.(int64))
		return true
	case reflect.Bool:
		field.SetBool(val.(bool))
		return true
	default:

		return false
	}

}

func GetAttr(o interface{}, attr string) interface{} {
	v := reflect.ValueOf(o).Elem()
	f := v.FieldByName(attr)
	return f
}

func HasAttr(o interface{}, attr string) bool {
	v := reflect.ValueOf(o)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.FieldByName(attr).IsValid()
}

func ReflectNew(target interface{}) interface{} {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	newStruct := reflect.New(t)
	return newStruct.Interface()
}

func Assert[T any](val any) (T, bool) {
	p, ok := val.(T)
	return p, ok
}
