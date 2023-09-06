# VMockHelper

This package contains some helper functions for writing tests using [gomock](https://github.com/golang/mock) mocks

## Functions:

### MockCallsAndPrintExpected

Takes a gomock interface and alias, automatically creates an expected call for all methods, 
and prints the inputs received when one is called. 

The alias is whatever variable name you used for your mock. It is used to print an exact expected mock.

Each mock call will return zero values by default.  You can optionally include a list of response arguments for the mock
call to return, but it will try to return those same arguments for every method, so this will not work in all cases.

#### Example Usage:

```
// Create gomock mock
mockLSP := lspSDK.NewMockListingSyncProServiceClientInterface(ctrl)

vmockhelper.MockCallsAndPrintExpected(mockLSP, "mockLSP")
```
Any calls to this service that happen when a test is run will automatically return zero values, and print an alert in 
the console that contains the exact expected call:
```
Alert                                vmockhelper/testgen.go:70   
m.lspClient.EXPECT().TriggerStatsCollection(gomock.Any(), &listing_sync_pro_v1.CollectStatsRequest{AccountGroupId:"AG-5VX5MZ2DQ4", PartnerId:"ABC", ServiceProvider:listing_sync_pro_v1.ServiceProvider(1), ServiceAreaBusiness:false}, []grpc.CallOption{}).Return(nil, nil)
```
This call represents what inputs the service was called with.  This information can be used help build test cases, or
identify when services are being called with unexpected parameters

### UseRealAndPrintExpected

Functions nearly identically to `MockCallsAndPrintExpected`, except every request that comes in is relayed to a real
instance of the service, and the response is printed along with the inputs.

This function can be used to make real requests to other microservices when your test is run, which is useful for building
test cases containing large amounts of real data the test writer does not want to fill in themselves.

Some care should be taken when using this, as it basically turns your test into an integration test.  Be sure to remove any
usages of this function before you push your code so that your builds dont end up trying to hit other services

Example usage:
```
ctx := context.Background()

// Set your VENDASTA_APPLICATION_CREDENTIALS_JSON and ENVIRONMENT environment variables
err := vmockhelper.GetCredentials(config.Demo) 
if err != nil {
	fmt.Println(err.Error())
	return
}

// Create a real instance of the account group SDK
agSDK, err := accountgroup.NewClient(ctx, config.Demo)
if err != nil {
	fmt.Println(err.Error())
	return
}

// Create a mock account group SDK
agMock := accountgroupwrapper.NewMockServer(ctrl)


(within test case)

vmockhelper.UseRealAndPrintExpected(agMock, agSDK, "agMock")
```
Result:
```
Alert                                vmockhelper/testgen.go:115  
agMock.EXPECT().Get(gomock.Any(), "AG-5VX5MZ2DQ4").Return(&accountgroup.AccountGroup{AccountGroupID:"AG-5VX5MZ2DQ4", Created:time.Date(2022, 3, 14, 19, 10, 30, 367037245, time.UTC), Updated:time.Date(2022, 6, 8, 0, 5, 7, 586542075, time.UTC), Deleted:time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), Version:0, NAPData:&accountgroup.NAPData{CompanyName:"SoZo Coffee House", Address:"1982 North Alma School Road", Address2:"", City:"Chandler", State:"AZ", Zip:"85224", Country:"US", Website:"http://www.sozocoffee.org/", WorkNumber:[]string{"4807267696"}, CallTrackingNumber:[]string{""}, Location:&accountgroup.Geo{Latitude:33.3334111, Longitude:-111.8607305}, Timezone:"America/Phoenix", FieldMask:nil, ServiceAreaBusiness:false}, ContactDetails:nil, Accounts:nil, ListingDistribution:nil, ListingSyncPro:nil, Associations:nil, ExternalIdentifiers:&accountgroup.ExternalIdentifiers{Origin:"partner-center-eac", JobID:nil, CustomerIdentifier:"", Tags:nil, ActionLists:nil, SocialProfileID:"SCP-2B6FC3260B8F414B9AF57CD0EC38418B", PartnerID:"ABC", MarketID:"default", TaxIDs:[]string{"food", "restaurants:cafes"}, SalesPersonID:"", AdditionalSalesPersonIDs:nil, SalesforceID:"", FieldMask:nil, VCategoryIDs:nil}, SocialURLs:nil, HoursOfOperation:nil, Snapshots:nil, LegacyProductDetails:nil, Status:nil, RichData:nil, Constraints:nil, AdditionalCompanyInfo:nil, MarketingInfo:nil}, nil)
```
A code representation of the response will be printed within the return of the expected mock call.  Now that you have real
data to work with you can easily copy and paste that data into a test case. 

### NOTE
The code printed from these functions represents the actual data the services received and returned during the test run.
It is still up to the whoever is writing the tests to check those inputs and outputs and make sure they are matching the
values we expect to see.

### Updating and publishing steps:
https://go.dev/doc/modules/publishing
