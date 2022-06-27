package vmockhelper

import (
	"context"
	"fmt"
	"github.com/vendasta/gosdks/config"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"github.com/golang/mock/gomock"
	"github.com/vendasta/gosdks/logging"
	"github.com/Short-Hop/vrender"
)

const mockFMT = "\n%s.EXPECT().%s(%s).Return(%s)\n"

// GetCredentials sets environment variables required to initialize microservice Go SDKs locally.  Should be called at
// the beginning of a test function before any test cases are run
func GetCredentials(env config.Env) error {
	err := os.Setenv("ENVIRONMENT", env.Name())
	if err != nil {
		return err
	}
	out, err := exec.Command("mscli", "auth", "path", "-e", env.Name(), "--skip-version-check").Output()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return os.Setenv("VENDASTA_APPLICATION_CREDENTIALS_JSON", string(out))
}

// MockCallsAndPrintExpected takes a gomock interface and alias, automatically creates an expected call for all methods,
// and prints the inputs received when one is called. Each mock call will return zero values by default.  You can
// optionally include a list of response arguments for the mock call to return, but it will try to return those same
// arguments for every method, so this will not work in all cases
func MockCallsAndPrintExpected(gomockObject interface{}, mockAlias string, mockResponseArgs ...interface{}) {
	mock := reflect.ValueOf(gomockObject)
	mockType := reflect.TypeOf(gomockObject)

	mockRecorder := mock.MethodByName("EXPECT").Call([]reflect.Value{})[0]

	var methodNames []string
	for i := 0; i < mockRecorder.Type().NumMethod(); i++ {
		methodNames = append(methodNames, mockRecorder.Type().Method(i).Name)
	}
	for _, methodName := range methodNames {
		methodType := mock.MethodByName(methodName).Type()
		method, _ := mockType.MethodByName(methodName)

		var inputParameters []reflect.Value
		numberOfInputParameters := mockRecorder.MethodByName(methodName).Type().NumIn()
		for i := 0; i < numberOfInputParameters; i++ {
			inputParameters = append(inputParameters, reflect.ValueOf(gomock.Any()))
		}

		mockCall := mockRecorder.MethodByName(methodName).Call(inputParameters)[0]
		call := mockCall.Interface().(*gomock.Call)
		doAndReturnFunction := reflect.MakeFunc(methodType, func(args []reflect.Value) []reflect.Value {
			var returns []reflect.Value
			for i := 0; i < methodType.NumOut(); i++ {
				if len(mockResponseArgs) > i && mockResponseArgs[i] != nil {
					returns = append(returns, reflect.ValueOf(mockResponseArgs[i]))
				} else {
					returns = append(returns, reflect.Zero(methodType.Out(i)))
				}
			}
			inputString := valuesToCodeString(args)
			returnString := valuesToCodeString(returns)

			logging.Alertf(context.Background(), fmt.Sprintf(mockFMT, mockAlias, method.Name, inputString, returnString))
			return returns
		})
		call.DoAndReturn(doAndReturnFunction.Interface()).AnyTimes()
	}
}

// UseRealAndPrintExpected takes in a gomock object and an instance of the actual service being mocked.  When a mock
// method is called, it calls the same method on the real service and prints the inputs and outputs
func UseRealAndPrintExpected(gomockObject interface{}, realService interface{}, mockAlias string) {
	mock := reflect.ValueOf(gomockObject)
	mockType := reflect.TypeOf(gomockObject)

	mockRecorder := mock.MethodByName("EXPECT").Call([]reflect.Value{})[0]

	var methodNames []string
	for i := 0; i < mockRecorder.Type().NumMethod(); i++ {
		methodNames = append(methodNames, mockRecorder.Type().Method(i).Name)
	}
	for _, methodName := range methodNames {
		methodType := mock.MethodByName(methodName).Type()
		method, _ := mockType.MethodByName(methodName)

		var inputParameters []reflect.Value
		numberOfInputParameters := mockRecorder.MethodByName(methodName).Type().NumIn()
		for i := 0; i < numberOfInputParameters; i++ {
			inputParameters = append(inputParameters, reflect.ValueOf(gomock.Any()))
		}

		mockCall := mockRecorder.MethodByName(methodName).Call(inputParameters)[0]
		call := mockCall.Interface().(*gomock.Call)
		function := reflect.MakeFunc(methodType, func(args []reflect.Value) []reflect.Value {
			v := reflect.ValueOf(realService)
			if v.MethodByName(method.Name).Type().IsVariadic() {
				lastArgument := args[len(args)-1]
				args = args[:len(args)-1]
				var variadicList []reflect.Value
				for i := 0; i < lastArgument.Len(); i++ {
					variadicList = append(variadicList, lastArgument.Index(i))
				}
				args = append(args, variadicList...)
			}
			returns := v.MethodByName(method.Name).Call(args)
			inputString := valuesToCodeString(args)
			returnString := valuesToCodeString(returns)

			logging.Alertf(context.Background(), fmt.Sprintf(mockFMT, mockAlias, method.Name, inputString, returnString))
			return returns
		})
		call.DoAndReturn(function.Interface()).AnyTimes()
	}
}

func valuesToCodeString(values []reflect.Value) string {
	var full string
	for _, value := range values {
		if isContext(value) {
			full += "gomock.Any(), "
			continue
		}
		full += vrender.Render(value.Interface()) + ", "
	}
	full = strings.TrimSuffix(full, ", ")
	return full
}

func isContext(v reflect.Value) bool {
	if v.Type().Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return true
	}
	if v.CanInterface() && v.Interface() != nil {
		return reflect.TypeOf(v.Interface()).Implements(reflect.TypeOf((*context.Context)(nil)).Elem())
	}
	return v.Type().Implements(reflect.TypeOf((*context.Context)(nil)).Elem())
}
