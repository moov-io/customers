# \AdminApi

All URIs are relative to *http://localhost:9097*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCustomerDisclaimer**](AdminApi.md#CreateCustomerDisclaimer) | **Post** /customers/{customerID}/disclaimers | Create a disclaimer for a customer to approve
[**GetVersion**](AdminApi.md#GetVersion) | **Get** /version | Show the current version
[**UpdateCustomerAddress**](AdminApi.md#UpdateCustomerAddress) | **Put** /customers/{customerID}/addresses/{addressID} | Update a Customers address
[**UpdateCustomerStatus**](AdminApi.md#UpdateCustomerStatus) | **Put** /customers/{customerID}/status | Update a Customers status and initiate required checks



## CreateCustomerDisclaimer

> CreateCustomerDisclaimer(ctx, customerID, createUserDisclaimer)

Create a disclaimer for a customer to approve

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

Show the current version

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

Update a Customers address

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

Update a Customers status and initiate required checks

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

