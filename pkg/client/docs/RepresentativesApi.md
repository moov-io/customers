# \RepresentativesApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddRepresentative**](RepresentativesApi.md#AddRepresentative) | **Post** /customers/{customerID}/representatives | Add Customer Representative
[**AddRepresentativeAddress**](RepresentativesApi.md#AddRepresentativeAddress) | **Post** /customers/{customerID}/representatives/{representativeID}/address | Add Customer Representative Address
[**DeleteRepresentative**](RepresentativesApi.md#DeleteRepresentative) | **Delete** /customers/{customerID}/representatives/{representativeID} | Delete Customer Representative
[**DeleteRepresentativeAddress**](RepresentativesApi.md#DeleteRepresentativeAddress) | **Delete** /customers/{customerID}/representatives/{representativeID}/addresses/{addressID} | Delete a Customer Representative Address
[**UpdateRepresentative**](RepresentativesApi.md#UpdateRepresentative) | **Put** /customers/{customerID}/representatives/{representativeID} | Update Customer Representative
[**UpdateRepresentativeAddress**](RepresentativesApi.md#UpdateRepresentativeAddress) | **Put** /customers/{customerID}/representatives/{representativeID}/addresses/{addressID} | Update Customer Representative Address



## AddRepresentative

> Customer AddRepresentative(ctx, xRequestID, customerID, createRepresentative, optional)

Add Customer Representative

Add a Customer Representative

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**xRequestID** | **string**| Optional requestID allows application developer to trace requests through the systems logs | 
**customerID** | **string**| customerID of the Customer for whom to add the representative | 
**createRepresentative** | [**CreateRepresentative**](CreateRepresentative.md)|  | 
 **optional** | ***AddRepresentativeOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AddRepresentativeOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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


## AddRepresentativeAddress

> Representative AddRepresentativeAddress(ctx, customerID, representativeID, createAddress, optional)

Add Customer Representative Address

Add an address to an existing customer representative record

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to add the address onto | 
**representativeID** | **string**| representativeID of the Customer representative for whom to add the address | 
**createAddress** | [**CreateAddress**](CreateAddress.md)|  | 
 **optional** | ***AddRepresentativeAddressOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AddRepresentativeAddressOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**Representative**](Representative.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRepresentative

> DeleteRepresentative(ctx, customerID, representativeID)

Delete Customer Representative

Deletes a Customer Representative

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**representativeID** | **string**| Representative ID | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRepresentativeAddress

> DeleteRepresentativeAddress(ctx, customerID, representativeID, addressID)

Delete a Customer Representative Address

Deletes a customer representative's address

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**representativeID** | **string**| Customer Representative ID | 
**addressID** | **string**| Address ID | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRepresentative

> UpdateRepresentative(ctx, customerID, representativeID, createRepresentative)

Update Customer Representative

Updates the specified Customer Representative

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**representativeID** | **string**| Representative ID | 
**createRepresentative** | [**CreateRepresentative**](CreateRepresentative.md)|  | 

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


## UpdateRepresentativeAddress

> UpdateRepresentativeAddress(ctx, customerID, representativeID, addressID, updateAddress)

Update Customer Representative Address

Updates the specified customer representative address

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**representativeID** | **string**| Customer Representative ID | 
**addressID** | **string**| Address ID | 
**updateAddress** | [**UpdateAddress**](UpdateAddress.md)|  | 

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

