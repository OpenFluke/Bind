package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"syscall/js"
	"time"

	"phase" // Adjust the import path if necessary
)

// methodWrapper dynamically wraps each method of Phase to expose it to JavaScript
func methodWrapper(bp *phase.Phase, methodName string) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		method := reflect.ValueOf(bp).MethodByName(methodName)
		if !method.IsValid() {
			return fmt.Sprintf("Method %s not found", methodName)
		}

		methodType := method.Type()
		if len(args) == 0 || args[0].String() == "" {
			if methodType.NumIn() == 0 {
				inputs := []reflect.Value{}
				results := method.Call(inputs)
				return serializeResults(results)
			}
			return "No arguments provided"
		}

		var params []interface{}
		paramJSON := args[0].String()
		if err := json.Unmarshal([]byte(paramJSON), &params); err != nil {
			return fmt.Sprintf("Invalid JSON input: %v", err)
		}

		expectedParams := methodType.NumIn()
		if len(params) != expectedParams {
			return fmt.Sprintf("Expected %d parameters, got %d", expectedParams, len(params))
		}

		inputs := make([]reflect.Value, expectedParams)
		for i := 0; i < expectedParams; i++ {
			param := params[i]
			expectedType := methodType.In(i)

			switch expectedType.Kind() {
			case reflect.Slice:
				if expectedType == reflect.TypeOf([]int{}) {
					if val, ok := param.([]interface{}); ok {
						slice := make([]int, len(val))
						for j, v := range val {
							if num, ok := v.(float64); ok { // JSON numbers are float64
								slice[j] = int(num)
							} else {
								return fmt.Sprintf("Parameter %d: invalid slice element type %T", i, v)
							}
						}
						inputs[i] = reflect.ValueOf(slice)
					} else {
						return fmt.Sprintf("Parameter %d: expected []int, got %T", i, param)
					}
				} else {
					return fmt.Sprintf("Parameter %d: unsupported slice type %s", i, expectedType.String())
				}
			case reflect.Map:
				if expectedType.Key().Kind() == reflect.Int && expectedType.Elem().Kind() == reflect.Float64 {
					jsonMap, ok := param.(map[string]interface{})
					if !ok {
						return fmt.Sprintf("Parameter %d: expected map[int]float64, got %T", i, param)
					}
					inputMap := make(map[int]float64)
					for keyStr, val := range jsonMap {
						key, err := strconv.Atoi(keyStr)
						if err != nil {
							return fmt.Sprintf("Parameter %d: invalid map key %s", i, keyStr)
						}
						value, ok := val.(float64)
						if !ok {
							return fmt.Sprintf("Parameter %d: invalid map value for key %s: %T", i, keyStr, val)
						}
						inputMap[key] = value
					}
					inputs[i] = reflect.ValueOf(inputMap)
				} else {
					return fmt.Sprintf("Parameter %d: unsupported map type %s", i, expectedType.String())
				}
			case reflect.Int:
				if val, ok := param.(float64); ok {
					inputs[i] = reflect.ValueOf(int(val))
				} else {
					return fmt.Sprintf("Parameter %d: expected int, got %T", i, param)
				}
			case reflect.Float64:
				if val, ok := param.(float64); ok {
					inputs[i] = reflect.ValueOf(val)
				} else {
					return fmt.Sprintf("Parameter %d: expected float64, got %T", i, param)
				}
			case reflect.Bool:
				if val, ok := param.(bool); ok {
					inputs[i] = reflect.ValueOf(val)
				} else {
					return fmt.Sprintf("Parameter %d: expected bool, got %T", i, param)
				}
			case reflect.String:
				if val, ok := param.(string); ok {
					inputs[i] = reflect.ValueOf(val)
				} else {
					return fmt.Sprintf("Parameter %d: expected string, got %T", i, param)
				}
			case reflect.TypeOf(time.Duration(0)).Kind():
				if val, ok := param.(float64); ok {
					inputs[i] = reflect.ValueOf(time.Duration(val))
				} else {
					return fmt.Sprintf("Parameter %d: expected duration, got %T", i, param)
				}
			default:
				inputs[i] = reflect.Zero(expectedType)
				if methodName != "GetphaseMethods" {
					return fmt.Sprintf("Parameter %d: unsupported type %s", i, expectedType.String())
				}
			}
		}

		results := method.Call(inputs)
		return serializeResults(results)
	})
}

// serializeResults converts method results to a JSON string
func serializeResults(results []reflect.Value) interface{} {
	if len(results) == 0 {
		return "[]" // Return empty array for void methods
	}

	output := make([]interface{}, len(results))
	for i, result := range results {
		output[i] = result.Interface()
	}

	resultJSON, err := json.Marshal(output)
	if err != nil {
		return fmt.Sprintf("Failed to marshal results: %v", err)
	}
	return string(resultJSON)
}

// newPhaseWrapper is a factory function that creates a new Phase instance
func newPhaseWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		bp := phase.NewPhase()
		obj := js.Global().Get("Object").New()
		methods, err := bp.GetphaseMethods()
		if err != nil {
			obj.Set("error", fmt.Sprintf("Error getting methods: %v", err))
			return obj
		}
		for _, method := range methods {
			obj.Set(method.MethodName, methodWrapper(bp, method.MethodName))
		}
		return obj
	})
}

func main() {
	js.Global().Set("NewPhase", newPhaseWrapper())
	select {}
}
