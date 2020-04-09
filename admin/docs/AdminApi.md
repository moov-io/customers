# \AdminApi

All URIs are relative to *http://localhost:9097*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCustomerDisclaimer**](AdminApi.md#CreateCustomerDisclaimer) | **Post** /customers/{customerID}/disclaimers | Create disclaimer
[**GetVersion**](AdminApi.md#GetVersion) | **Get** /version | Get Version
[**UpdateCustomerAddress**](AdminApi.md#UpdateCustomerAddress) | **Put** /customers/{customerID}/addresses/{addressID} | Update customers address
[**UpdateCustomerStatus**](AdminApi.md#UpdateCustomerStatus) | **Put** /customers/{customerID}/status | Update Customer status



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


## GetVersion

> string GetVersion(ctx, )

Get Version

Show the current version of Customers

### Required Parameters

This endpoint does not need any parameter.

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateCustomerAddress

> UpdateCustomerAddress(ctx, customerID, addressID, updateCustomerAddress)

Update customers address

Updates the specified customer address

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**addressID** | **string**| Address ID | 
**updateCustomerAddress** | [**UpdateCustomerAddress**](UpdateCustomerAddress.md)|  | 

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


## UpdateCustomerStatus

> UpdateCustomerStatus(ctx, customerID, updateCustomerStatus)

Update Customer status

Updates a customer status and initiates the required checks for that new status

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**updateCustomerStatus** | [**UpdateCustomerStatus**](UpdateCustomerStatus.md)|  | 

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

