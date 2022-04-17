package sync

import (
	"reflect"
)

// Clear garbage collectable struct
func clear(v interface{}) {
	// https://stackoverflow.com/questions/29168905/how-to-clear-values-of-a-instance-of-a-type-struct-dynamically/51006888#51006888
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}
