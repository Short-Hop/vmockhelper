package vmockhelper

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/short-hop/vrender"
	"github.com/vendasta/gosdks/logging"
)

// Record will begin to record mock calls
func Record() {
	record = true
}

// Clear will clear all recorded mock calls
func Clear() {
	record = false
	recordedCalls = []Call{}
}

var record bool
var recordedCalls []Call

type Call struct {
	method  string
	alias   string
	args    []reflect.Value
	returns []reflect.Value
}

// PrintTestCase takes recorded calls and prints a test case that can be used to test the same functionality
func PrintTestCase() {
	template := `type testCase struct {
{{caseType}}}
cases := []*testCase{
{{cases}}
}`
	template = strings.Replace(template, "{{caseType}}", generateTestCaseType(), -1)
	template = strings.Replace(template, "{{cases}}", generateTestCase(), -1)
	logging.Alertf(context.Background(), template)
}

func generateTestCaseType() string {
	caseType := ""
	for _, call := range recordedCalls {
		indexOffset := 1
		for i, arg := range call.args {
			if isContext(arg) {
				indexOffset--
				continue
			}
			caseType += fmt.Sprintf("\t%s%sIn%d %s\n", call.alias, call.method, i+indexOffset, arg.Type().String())
		}
		indexOffset = 1
		for i, arg := range call.returns {
			if isContext(arg) {
				indexOffset--
				continue
			}
			caseType += fmt.Sprintf("\t%s%sOut%d %s\n", call.alias, call.method, i+indexOffset, arg.Type().String())
		}
	}
	return caseType
}

func generateTestCase() string {
	testCase := "{\n"
	for _, call := range recordedCalls {
		indexOffset := 1
		for i, arg := range call.args {
			if isContext(arg) {
				indexOffset--
				continue
			}
			testCase += fmt.Sprintf("\t%s%sIn%d: %s,\n", call.alias, call.method, i+indexOffset, vrender.Render(arg.Interface()))
		}
		indexOffset = 1
		for i, arg := range call.returns {
			if isContext(arg) {
				indexOffset--
				continue
			}
			testCase += fmt.Sprintf("\t%s%sOut%d: %s,\n", call.alias, call.method, i+indexOffset, vrender.Render(arg.Interface()))
		}
	}
	testCase += fmt.Sprintln("},")
	return testCase
}
