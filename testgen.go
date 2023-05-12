package vmockhelper

import (
	"fmt"
	"reflect"
	"strings"
)

// GenerateTestTemplate generates a test template for a given service and method
func GenerateTestTemplate(s interface{}, methodName string) {
	var mocks []string
	var mockHelpers []string
	var fields []string

	sType := reflect.TypeOf(s)
	//packageName = sType.PkgPath()

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
	t = strings.ReplaceAll(t, "{{mockHelpers}}", strings.Join(mockHelpers, "\n"))
	t = strings.ReplaceAll(t, "{{fields}}", strings.Join(fields, "\n"))
	t = strings.ReplaceAll(t, "{{methodName}}", methodName)
	t = strings.ReplaceAll(t, "{{serviceType}}", sType.Name())
	fmt.Println(t)
}
