package reflection_tools

import "reflect"

func GetFieldNameByValue(instance interface{}, value interface{}) string {
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ""
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if reflect.DeepEqual(field.Interface(), value) {
			return val.Type().Field(i).Name
		}
	}
	return ""
}
