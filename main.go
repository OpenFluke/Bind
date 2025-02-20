package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"syscall/js"
	"time"

	"phase" // Adjust the import path if necessary
)

// methodWrapper dynamically wraps each method of Blueprint to expose it to JavaScript
func methodWrapper(bp *phase.Phase, methodName string) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		method := reflect.ValueOf(bp).MethodByName(methodName)
		if !method.IsValid() {
			return fmt.Sprintf("Method %s not found", methodName)
		}

		methodType := method.Type()
		var inputs []reflect.Value

		// Check for no parameters
		if len(args) == 0 && methodType.NumIn() == 0 {
			inputs = []reflect.Value{}
		} else if len(args) > 0 {
			// If parameters are expected, parse the JSON input
			var paramValues []interface{}
			paramJSON := args[0].String()

			if err := json.Unmarshal([]byte(paramJSON), &paramValues); err != nil {
				return fmt.Sprintf("Invalid JSON input: %v", err)
			}

			// Prepare inputs based on method's expected parameter types
			for i, param := range paramValues {
				expectedType := methodType.In(i)

				switch expectedType.Kind() {
				case reflect.Int:
					inputs = append(inputs, reflect.ValueOf(int(param.(float64))))
				case reflect.Float64:
					inputs = append(inputs, reflect.ValueOf(param.(float64)))
				case reflect.Bool:
					inputs = append(inputs, reflect.ValueOf(param.(bool)))
				case reflect.String:
					inputs = append(inputs, reflect.ValueOf(param.(string)))
				case reflect.TypeOf(time.Duration(0)).Kind():
					inputs = append(inputs, reflect.ValueOf(time.Duration(param.(float64))))
				default:
					inputs = append(inputs, reflect.Zero(expectedType))
				}
			}
		}

		// Call the method and capture results
		results := method.Call(inputs)

		// Format the results as JSON for output
		output := make([]interface{}, len(results))
		for i, result := range results {
			output[i] = result.Interface()
		}
		resultJSON, _ := json.Marshal(output)
		return string(resultJSON)
	})
}

func main() {
	// Create an instance of Blueprint
	bp := phase.NewPhase()

	// Retrieve all methods using reflection from introspection
	methods, err := bp.GetphaseMethods()
	if err != nil {
		fmt.Println("Error getting methods:", err)
		return
	}

	// Expose each method as a JavaScript function
	for _, method := range methods {
		js.Global().Set(method.MethodName, methodWrapper(bp, method.MethodName))
	}

	// Keep the WebAssembly program running
	select {}
}
