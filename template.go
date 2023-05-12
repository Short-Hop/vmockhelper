package vmockhelper

const template = `
func Test_{{methodName}}(t *testing.T) {
	type testCase struct {
		name             string
		expectedResponse interface{}
		expectedErr      error
	}
	cases := []*testCase{
		{},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)

			vmockhelper.Record()

			{{mocks}}

			{{mockHelpers}}

			s := {{serviceType}}{
				{{fields}}
			}

			resp, err := s.{{methodName}}() // TODO: set args

			vmockhelper.Record()

			assert.Equal(t, c.expectedResponse, resp)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}`

const mockTemplate = `mock{{fieldName}} := {{interfacePackage}}NewMock{{interfaceName}}(ctrl)`
const mockHelperTemplate = `vmockhelper.MockCallsAndPrintExpected(mock{{fieldName}}, "mock{{fieldName}}")`
const fieldTemplate = `{{fieldName}}: mock{{fieldName}},`
