# \CustomersApi

All URIs are relative to *http://localhost:9097*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCustomerDisclaimer**](CustomersApi.md#CreateCustomerDisclaimer) | **Post** /customers/{customerID}/disclaimers | Create disclaimer



## CreateCustomerDisclaimer

> CreateCustomerDisclaimer(ctx, customerID, createUserDisclaimer)

Create disclaimer

Create a disclaimer for the specified customerID to approve

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**createUserDisclaimer** | [**CreateUserDisclaimer**](CreateUserDisclaimer.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

