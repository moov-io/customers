# \CustomersApi

All URIs are relative to *http://localhost:9097*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCustomerDisclaimer**](CustomersApi.md#CreateCustomerDisclaimer) | **Post** /customers/{customerID}/disclaimers | Create disclaimer
[**UpdateAccountStatus**](CustomersApi.md#UpdateAccountStatus) | **Put** /customers/{customerID}/accounts/{accountID}/status | Update Account Status
[**UpdateCustomerAddress**](CustomersApi.md#UpdateCustomerAddress) | **Put** /customers/{customerID}/addresses/{addressID} | Update customers address
[**UpdateCustomerStatus**](CustomersApi.md#UpdateCustomerStatus) | **Put** /customers/{customerID}/status | Update Customer status



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


## UpdateAccountStatus

> UpdateAccountStatus(ctx, customerID, accountID, updateAccountStatus)

Update Account Status

Update the status for the specified accountID

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**accountID** | **string**| accountID of the Account to validate | 
**updateAccountStatus** | [**UpdateAccountStatus**](UpdateAccountStatus.md)|  | 

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

