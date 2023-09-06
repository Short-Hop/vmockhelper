package vmockhelper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/short-hop/vrender"
)

type testType struct {
	name      string
	valueType string
	value     interface{}
}

func (t testType) NameAndTypeString() string {
	return fmt.Sprintf("%s %s", t.name, t.valueType)
}

func (t testType) AssignedZeroValueString() string {
	return fmt.Sprintf("%s: %s", t.name, vrender.Render(t.value))
}

// GenerateTestTemplate generates a test template for a given service and method
func GenerateTestTemplate(s interface{}, methodName string) {
	var mocks []string
	var mockHelpers []string
	var fields []string

	sType := reflect.TypeOf(s)
	//packageName = sType.PkgPath()

	method, found := sType.MethodByName(methodName)
	if !found {
		panic("Method not found")
	}

	var inputNames []string
	var outputNames []string
	numberOfInputParameters := method.Type.NumIn()
	var inputs []testType
	for i := 1; i < numberOfInputParameters; i++ {
		//v := reflect.New(method.Type.In(i)).Elem()
		//v.SetZero()
		inputs = append(inputs, testType{
			name:      fmt.Sprintf("%sInput%d", methodName, i),
			valueType: method.Type.In(i).String(),
			value:     reflect.New(method.Type.In(i)).Elem().Interface(),
		})
	}

	var inputTypes string
	var testValues string
	for _, in := range inputs {
		inputTypes += fmt.Sprintf("%s\n", in.NameAndTypeString())
		inputNames = append(inputNames, fmt.Sprintf("c.%s", in.name))
		testValues += fmt.Sprintf("%s,\n", in.AssignedZeroValueString())
	}

	var outputTypes string
	var asserts string
	var outputs []testType
	numberOfOutputParameters := method.Type.NumOut()
	for i := 0; i < numberOfOutputParameters; i++ {
		outputs = append(outputs, testType{
			name:      fmt.Sprintf("expectedOut%d", i+1),
			valueType: method.Type.Out(i).String(),
			value:     reflect.New(method.Type.Out(i)).Elem().Interface(),
		})
		outputNames = append(outputNames, fmt.Sprintf("out%d", i+1))
		asserts += fmt.Sprintf("assert.Equal(t, c.expectedOut%d, out%d)\n", i+1, i+1)
	}

	for _, out := range outputs {
		outputTypes += fmt.Sprintf("%s\n", out.NameAndTypeString())
		testValues += fmt.Sprintf("%s,\n", out.AssignedZeroValueString())
	}

	outputTypes = strings.TrimSuffix(outputTypes, "\n")
	inputTypes = strings.TrimSuffix(inputTypes, "\n")
	asserts = strings.TrimSuffix(asserts, "\n")

	sType = sType.Elem()
	for i := 0; i < sType.NumField(); i++ {
		field := sType.Field(i)
		fieldEntry := strings.ReplaceAll(fieldTemplate, "{{fieldName}}", field.Name)
		fieldEntry = strings.ReplaceAll(fieldEntry, "{{interfaceName}}", field.Type.Name())
		fields = append(fields, fieldEntry)

		if len(strings.Split(field.Type.String(), ".")) > 1 {
			interfaceName := strings.Split(field.Type.String(), ".")[1]
			interfacePackageName := strings.Split(field.Type.String(), ".")[0]

			mockEntry := strings.ReplaceAll(mockTemplate, "{{interfaceName}}", interfaceName)
			mockEntry = strings.ReplaceAll(mockEntry, "{{fieldName}}", field.Name)
			mockEntry = strings.ReplaceAll(mockEntry, "{{interfacePackage}}", interfacePackageName+".")
			mocks = append(mocks, mockEntry)
		} else {
			mockEntry := strings.ReplaceAll(mockTemplate, "{{interfaceName}}", field.Type.Name())
			mockEntry = strings.ReplaceAll(mockEntry, "{{fieldName}}", field.Name)
			mockEntry = strings.ReplaceAll(mockEntry, "{{interfacePackage}}", "")
			mocks = append(mocks, mockEntry)
		}
		mockHelpers = append(mockHelpers, strings.ReplaceAll(mockHelperTemplate, "{{fieldName}}", field.Name))
	}

	t := strings.ReplaceAll(template, "{{mocks}}", strings.Join(mocks, "\n"))
	t = strings.ReplaceAll(t, "{{inputs}}", strings.Join(inputNames, ", "))
	t = strings.ReplaceAll(t, "{{outputTypes}}", outputTypes)
	t = strings.ReplaceAll(t, "{{inputTypes}}", inputTypes)
	t = strings.ReplaceAll(t, "{{outputs}}", strings.Join(outputNames, ", "))
	t = strings.ReplaceAll(t, "{{mockHelpers}}", strings.Join(mockHelpers, "\n"))
	t = strings.ReplaceAll(t, "{{fields}}", strings.Join(fields, "\n"))
	t = strings.ReplaceAll(t, "{{methodName}}", methodName)
	t = strings.ReplaceAll(t, "{{serviceType}}", sType.Name())
	t = strings.ReplaceAll(t, "{{asserts}}", asserts)
	t = strings.ReplaceAll(t, "{{testValues}}", testValues)
	fmt.Println(t)

	//err := os.WriteFile("generated_test.go", []byte(t), 0644)
	//if err != nil {
	//	panic(err)
	//}
}
