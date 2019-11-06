# \CustomersApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AcceptDisclaimer**](CustomersApi.md#AcceptDisclaimer) | **Post** /customers/{customerID}/disclaimers/{disclaimerID} | Accept a disclaimer for the given customer
[**AddCustomerAddress**](CustomersApi.md#AddCustomerAddress) | **Post** /customers/{customerID}/address | Add an Address onto an existing Customer record
[**CreateCustomer**](CustomersApi.md#CreateCustomer) | **Post** /customers | Create a new customer
[**GetCustomer**](CustomersApi.md#GetCustomer) | **Get** /customers/{customerID} | Retrieves a Customer object associated with the customer ID.
[**GetCustomerDisclaimers**](CustomersApi.md#GetCustomerDisclaimers) | **Get** /customers/{customerID}/disclaimers | Get active disclaimers for the given customer
[**GetCustomerDocumentContents**](CustomersApi.md#GetCustomerDocumentContents) | **Get** /customers/{customerID}/documents/{documentID} | Retrieve the referenced document
[**GetCustomerDocuments**](CustomersApi.md#GetCustomerDocuments) | **Get** /customers/{customerID}/documents | Get documents for a customer
[**GetLatestOFACSearch**](CustomersApi.md#GetLatestOFACSearch) | **Get** /customers/{customerID}/ofac | Get the latest OFAC search for a customer
[**Ping**](CustomersApi.md#Ping) | **Get** /ping | Ping the Customers service to check if running
[**RefreshOFACSearch**](CustomersApi.md#RefreshOFACSearch) | **Put** /customers/{customerID}/refresh/ofac | Refresh OFAC search for a given Customer
[**ReplaceCustomerMetadata**](CustomersApi.md#ReplaceCustomerMetadata) | **Put** /customers/{customerID}/metadata | Replace the metadata object for a customer. Metadata is a map of unique keys associated to values to act as foreign key relationships or arbitrary data associated to a Customer.
[**UpdateCustomerStatus**](CustomersApi.md#UpdateCustomerStatus) | **Put** /customers/{customerID}/status | Update the status for a customer, which can only be updated by authenticated users with permissions.
[**UploadCustomerDocument**](CustomersApi.md#UploadCustomerDocument) | **Post** /customers/{customerID}/documents | Upload a document for the given customer.



## AcceptDisclaimer

> Disclaimer AcceptDisclaimer(ctx, customerID, disclaimerID, optional)

Accept a disclaimer for the given customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**disclaimerID** | **string**| Disclaimer ID | 
 **optional** | ***AcceptDisclaimerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AcceptDisclaimerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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


## AddCustomerAddress

> Customer AddCustomerAddress(ctx, customerID, createAddress, optional)

Add an Address onto an existing Customer record

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**createAddress** | [**CreateAddress**](CreateAddress.md)|  | 
 **optional** | ***AddCustomerAddressOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AddCustomerAddressOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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

 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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

> Customer GetCustomer(ctx, customerID, optional)

Retrieves a Customer object associated with the customer ID.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
 **optional** | ***GetCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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


## GetCustomerDisclaimers

> []Disclaimer GetCustomerDisclaimers(ctx, customerID, optional)

Get active disclaimers for the given customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
 **optional** | ***GetCustomerDisclaimersOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerDisclaimersOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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


## GetCustomerDocumentContents

> *os.File GetCustomerDocumentContents(ctx, customerID, documentID, optional)

Retrieve the referenced document

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**documentID** | **string**| Document ID | 
 **optional** | ***GetCustomerDocumentContentsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerDocumentContentsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

### Return type

[***os.File**](*os.File.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/pdf, image/_*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCustomerDocuments

> []Document GetCustomerDocuments(ctx, customerID, optional)

Get documents for a customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
 **optional** | ***GetCustomerDocumentsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerDocumentsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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


## GetLatestOFACSearch

> OfacSearch GetLatestOFACSearch(ctx, customerID, optional)

Get the latest OFAC search for a customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
 **optional** | ***GetLatestOFACSearchOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetLatestOFACSearchOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

### Return type

[**OfacSearch**](OFACSearch.md)

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


## RefreshOFACSearch

> OfacSearch RefreshOFACSearch(ctx, customerID, optional)

Refresh OFAC search for a given Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
 **optional** | ***RefreshOFACSearchOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a RefreshOFACSearchOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

### Return type

[**OfacSearch**](OFACSearch.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ReplaceCustomerMetadata

> Customer ReplaceCustomerMetadata(ctx, customerID, customerMetadata, optional)

Replace the metadata object for a customer. Metadata is a map of unique keys associated to values to act as foreign key relationships or arbitrary data associated to a Customer.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**customerMetadata** | [**CustomerMetadata**](CustomerMetadata.md)|  | 
 **optional** | ***ReplaceCustomerMetadataOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a ReplaceCustomerMetadataOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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


## UpdateCustomerStatus

> Customer UpdateCustomerStatus(ctx, customerID, updateCustomerStatus, optional)

Update the status for a customer, which can only be updated by authenticated users with permissions.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**updateCustomerStatus** | [**UpdateCustomerStatus**](UpdateCustomerStatus.md)|  | 
 **optional** | ***UpdateCustomerStatusOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UpdateCustomerStatusOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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


## UploadCustomerDocument

> Document UploadCustomerDocument(ctx, customerID, type_, file, optional)

Upload a document for the given customer.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**type_** | **string**| Document type (see Document type for values) | 
**file** | ***os.File*****os.File**| Document to be uploaded | 
 **optional** | ***UploadCustomerDocumentOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UploadCustomerDocumentOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 
 **xUserID** | **optional.String**| Moov User ID | 

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

