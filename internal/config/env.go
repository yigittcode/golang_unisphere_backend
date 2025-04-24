package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

// processStructFields walks through struct fields to override config with env vars
func processStructFields(s interface{}) error {
	val := reflect.ValueOf(s)

	// If pointer, get the underlying element
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Only process structs
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	// Iterate through all fields of the struct
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check if field is a struct, if so recursively process it
		if field.Kind() == reflect.Struct {
			err := processStructFields(field.Addr().Interface())
			if err != nil {
				return err
			}
			continue
		}

		// Get env tag from field
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue // Skip if no env tag
		}

		// Check if environment variable exists
		envValue, exists := os.LookupEnv(envTag)
		if !exists {
			continue // Skip if environment variable is not set
		}

		// Set field based on its type
		if err := setFieldFromEnv(field, envValue); err != nil {
			return fmt.Errorf("failed to set field %s from env var %s: %w", fieldType.Name, envTag, err)
		}
	}

	return nil
}

// setFieldFromEnv sets a field value from an environment variable string
func setFieldFromEnv(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle special case for time.Duration
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration format: %w", err)
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			// Handle regular integers
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer format: %w", err)
			}
			field.SetInt(intValue)
		}

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean format: %w", err)
		}
		field.SetBool(boolValue)

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float format: %w", err)
		}
		field.SetFloat(floatValue)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}
