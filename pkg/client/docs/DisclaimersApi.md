# \DisclaimersApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AcceptDisclaimer**](DisclaimersApi.md#AcceptDisclaimer) | **Post** /customers/{customerID}/disclaimers/{disclaimerID} | Accept Customer Disclaimer
[**GetCustomerDisclaimers**](DisclaimersApi.md#GetCustomerDisclaimers) | **Get** /customers/{customerID}/disclaimers | Get Customer Disclaimers



## AcceptDisclaimer

> Disclaimer AcceptDisclaimer(ctx, customerID, disclaimerID, optional)

Accept Customer Disclaimer

Accept a disclaimer for the given customer which could include a document also

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to accept a disclaimer | 
**disclaimerID** | **string**| disclaimerID of the Disclaimer to accept | 
 **optional** | ***AcceptDisclaimerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AcceptDisclaimerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**Disclaimer**](Disclaimer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCustomerDisclaimers

> []Disclaimer GetCustomerDisclaimers(ctx, customerID, optional)

Get Customer Disclaimers

Get active disclaimers for the given customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to get disclaimers | 
 **optional** | ***GetCustomerDisclaimersOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerDisclaimersOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**[]Disclaimer**](Disclaimer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

