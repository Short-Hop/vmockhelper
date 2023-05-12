package vrender

import (
	"fmt"
	"math/rand"
	"reflect"
)

// Fill fills in values with default or random values
func Fill(data interface{}, names ...string) {
	name := ""
	if len(names) > 0 {
		name = names[0]
	}

	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return // Skip if the value is not a pointer or is nil
	}

	value = value.Elem() // Get the underlying value of the pointer

	switch value.Kind() {
	case reflect.Struct:
		numFields := value.NumField()
		for i := 0; i < numFields; i++ {
			field := value.Field(i)
			if field.CanSet() {
				Fill(field.Addr().Interface(), value.Type().Field(i).Name) // Recursively call the function for nested structs
			}
		}
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			Fill(value.Index(i).Addr().Interface(), fmt.Sprintf("%s%d", name, i)) // Recursively call the function for each element in the slice
		}
	case reflect.Map:
		keys := value.MapKeys()
		for _, key := range keys {
			Fill(value.MapIndex(key).Addr().Interface(), key.String()) // Recursively call the function for each value in the map
		}
	case reflect.Ptr:
		if value.IsNil() {
			newValue := reflect.New(value.Type().Elem())
			value.Set(newValue)
		}
		Fill(value.Interface(), name) // Recursively call the function for the underlying value of the pointer
	case reflect.String:
		value.SetString(name) // Set random string of length 10
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value.SetInt(int64(rand.Intn(100))) // Set random integer value between 0 and 99
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value.SetUint(uint64(rand.Intn(100))) // Set random unsigned integer value between 0 and 99
	case reflect.Float32, reflect.Float64:
		value.SetFloat(float64(rand.Float32() * 100.0)) // Set random float value between 0 and 99.9999
	case reflect.Bool:
		value.SetBool(rand.Float32() < 0.5) // Set random boolean value
	}

}
