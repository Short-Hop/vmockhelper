package vmockhelper

const template = `
func Test_{{methodName}}(t *testing.T) {
	type testCase struct {
		name             string
        {{inputTypes}}
		{{outputTypes}}
	}
	cases := []*testCase{
		{
			name: "Test {{methodName}}",
            {{testValues}}
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)

			{{mocks}}

			{{mockHelpers}}

			s := {{serviceType}}{
				{{fields}}
			}

			{{outputs}} := s.{{methodName}}({{inputs}})

			{{asserts}}
		})
	}
}`

const mockTemplate = `mock{{fieldName}} := {{interfacePackage}}NewMock{{interfaceName}}(ctrl)`
const mockHelperTemplate = `vmockhelper.MockCallsAndPrintExpected(mock{{fieldName}}, "mock{{fieldName}}")`
const fieldTemplate = `{{fieldName}}: mock{{fieldName}},`
