# \CustomersApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCustomer**](CustomersApi.md#CreateCustomer) | **Post** /customers | Create a new customer
[**GetCustomer**](CustomersApi.md#GetCustomer) | **Get** /customers/{customer_id} | Retrieves a Customer object associated with the customer ID.
[**GetCustomerDocuments**](CustomersApi.md#GetCustomerDocuments) | **Get** /customers/{customer_id}/documents | Get documents for a customer
[**Ping**](CustomersApi.md#Ping) | **Get** /ping | Ping the Customers service to check if running
[**UploadCustomerDocument**](CustomersApi.md#UploadCustomerDocument) | **Post** /customers/{customer_id}/documents | Upload a document for the given customer.



## CreateCustomer

> Customer CreateCustomer(ctx, createCustomer, optional)
Create a new customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**createCustomer** | [**CreateCustomer**](CreateCustomer.md)|  | 
 **optional** | ***CreateCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a CreateCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserId** | **optional.String**| Moov User ID header, required in all requests | 

### Return type

[**Customer**](Customer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCustomer

> Customer GetCustomer(ctx, customerId, optional)
Retrieves a Customer object associated with the customer ID.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerId** | **string**| Customer Id | 
 **optional** | ***GetCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserId** | **optional.String**| Moov User ID header, required in all requests | 

### Return type

[**Customer**](Customer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCustomerDocuments

> []Document GetCustomerDocuments(ctx, customerId, optional)
Get documents for a customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerId** | **string**| Customer Id | 
 **optional** | ***GetCustomerDocumentsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerDocumentsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserId** | **optional.String**| Moov User ID header, required in all requests | 

### Return type

[**[]Document**](Document.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Ping

> Ping(ctx, )
Ping the Customers service to check if running

### Required Parameters

This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadCustomerDocument

> Document UploadCustomerDocument(ctx, customerId, type_, file, optional)
Upload a document for the given customer.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerId** | **string**| Customer Id | 
**type_** | **string**| Document type (see Document type for values) | 
**file** | ***os.File*****os.File**| Document to be uploaded | 
 **optional** | ***UploadCustomerDocumentOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UploadCustomerDocumentOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserId** | **optional.String**| Moov User ID header, required in all requests | 

### Return type

[**Document**](Document.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

