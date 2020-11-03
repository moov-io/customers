# \AccountValidationApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CompleteAccountValidation**](AccountValidationApi.md#CompleteAccountValidation) | **Put** /customers/{customerID}/accounts/{accountID}/validations | Complete Account Validation
[**GetAccountValidation**](AccountValidationApi.md#GetAccountValidation) | **Get** /customers/{customerID}/accounts/{accountID}/validations/{validationID} | Get Account Validation
[**InitAccountValidation**](AccountValidationApi.md#InitAccountValidation) | **Post** /customers/{customerID}/accounts/{accountID}/validations | Initiate Account Validation



## CompleteAccountValidation

> CompleteAccountValidationResponse CompleteAccountValidation(ctx, customerID, accountID, completeAccountValidationRequest, optional)

Complete Account Validation

Complete account validation with specified strategy and vendor. 

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer the accountID belongs to | 
**accountID** | **string**| accountID of the Account to validate | 
**completeAccountValidationRequest** | [**CompleteAccountValidationRequest**](CompleteAccountValidationRequest.md)|  | 
 **optional** | ***CompleteAccountValidationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a CompleteAccountValidationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**CompleteAccountValidationResponse**](CompleteAccountValidationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAccountValidation

> AccountValidationResponse GetAccountValidation(ctx, customerID, accountID, validationID, optional)

Get Account Validation

Get information about account validation strategy, status, etc. 

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer the accountID belongs to | 
**accountID** | **string**| accountID of the Account to validate | 
**validationID** | **string**| ID of the Validation | 
 **optional** | ***GetAccountValidationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetAccountValidationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**AccountValidationResponse**](AccountValidationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitAccountValidation

> InitAccountValidationResponse InitAccountValidation(ctx, customerID, accountID, initAccountValidationRequest, optional)

Initiate Account Validation

Initiate account validation with specified strategy and vendor. 

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer the accountID belongs to | 
**accountID** | **string**| accountID of the Account to validate | 
**initAccountValidationRequest** | [**InitAccountValidationRequest**](InitAccountValidationRequest.md)|  | 
 **optional** | ***InitAccountValidationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a InitAccountValidationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**InitAccountValidationResponse**](InitAccountValidationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

