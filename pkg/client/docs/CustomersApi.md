# \CustomersApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AcceptDisclaimer**](CustomersApi.md#AcceptDisclaimer) | **Post** /customers/{customerID}/disclaimers/{disclaimerID} | Accept Customer Disclaimer
[**AddAddress**](CustomersApi.md#AddAddress) | **Post** /customers/{customerID}/address | Add customer address
[**AddCustomerRepresentative**](CustomersApi.md#AddCustomerRepresentative) | **Post** /customers/{customerID}/representatives | Add customer representative
[**AddCustomerRepresentativeAddress**](CustomersApi.md#AddCustomerRepresentativeAddress) | **Post** /customers/{customerID}/representatives/{representativeID}/address | Add customer representative address
[**CompleteAccountValidation**](CustomersApi.md#CompleteAccountValidation) | **Put** /customers/{customerID}/accounts/{accountID}/validations | Complete Account Validation
[**CreateCustomer**](CustomersApi.md#CreateCustomer) | **Post** /customers | Create Customer
[**DeleteCustomer**](CustomersApi.md#DeleteCustomer) | **Delete** /customers/{customerID} | Delete Customer
[**DeleteCustomerAccount**](CustomersApi.md#DeleteCustomerAccount) | **Delete** /customers/{customerID}/accounts/{accountID} | Delete Customer Account
[**DeleteCustomerDocument**](CustomersApi.md#DeleteCustomerDocument) | **Delete** /customers/{customerID}/documents/{documentID} | Delete Customer Document
[**DeleteAddress**](CustomersApi.md#DeleteAddress) | **Delete** /customers/{customerID}/addresses/{addressID} | Delete a customer&#39;s address
[**DeleteCustomerRepresentative**](CustomersApi.md#DeleteCustomerRepresentative) | **Delete** /customers/{customerID}/representatives/{representativeID} | Delete a customer&#39;s representative
[**DeleteCustomerRepresentativeAddress**](CustomersApi.md#DeleteCustomerRepresentativeAddress) | **Delete** /customers/{customerID}/representatives/{representativeID}/addresses/{addressID} | Delete a customer representative&#39;s address
[**GetAccountValidation**](CustomersApi.md#GetAccountValidation) | **Get** /customers/{customerID}/accounts/{accountID}/validations/{validationID} | Get Account Validation
[**GetCustomer**](CustomersApi.md#GetCustomer) | **Get** /customers/{customerID} | Get Customer
[**GetLatestOFACSearch**](CustomersApi.md#GetLatestOFACSearch) | **Get** /customers/{customerID}/ofac | Latest Customer OFAC search
[**Ping**](CustomersApi.md#Ping) | **Get** /ping | Ping Customers Service
[**RefreshOFACSearch**](CustomersApi.md#RefreshOFACSearch) | **Put** /customers/{customerID}/refresh/ofac | Refresh Customer OFAC search
[**ReplaceCustomerMetadata**](CustomersApi.md#ReplaceCustomerMetadata) | **Put** /customers/{customerID}/metadata | Update Customer Metadata
[**SearchCustomers**](CustomersApi.md#SearchCustomers) | **Get** /customers | Search Customers
[**UpdateCustomer**](CustomersApi.md#UpdateCustomer) | **Put** /customers/{customerID} | Update Customer
[**UpdateCustomerStatus**](CustomersApi.md#UpdateCustomerStatus) | **Put** /customers/{customerID}/status | Update Customer Status
[**UploadCustomerDocument**](CustomersApi.md#UploadCustomerDocument) | **Post** /customers/{customerID}/documents | Upload Customer Document
[**UpdateAddress**](CustomersApi.md#UpdateAddress) | **Put** /customers/{customerID}/addresses/{addressID} | Update customer&#39;s address
[**UpdateCustomerRepresentative**](CustomersApi.md#UpdateCustomerRepresentative) | **Put** /customers/{customerID}/representatives/{representativeID} | Update customer representative
[**UpdateCustomerRepresentativeAddress**](CustomersApi.md#UpdateCustomerRepresentativeAddress) | **Put** /customers/{customerID}/representatives/{representativeID}/addresses/{addressID} | Update customer representative&#39;s address



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


## AddAddress

> Customer AddAddress(ctx, customerID, createAddress, optional)

Add Customer Address

Add an Address onto an existing Customer record

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to add the address onto | 
**createAddress** | [**CreateAddress**](CreateAddress.md)|  | 
 **optional** | ***AddAddressOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AddAddressOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
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


## AddCustomerRepresentative

> Customer AddCustomerRepresentative(ctx, customerID, createCustomerRepresentative, optional)

Add customer representative

Add a Customer representative

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer for whom to add the representative | 
**createCustomerRepresentative** | [**CreateCustomerRepresentative**](CreateCustomerRepresentative.md)|  | 
 **optional** | ***AddCustomerRepresentativeOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AddCustomerRepresentativeOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
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


## AddCustomerRepresentativeAddress

> CustomerRepresentative AddCustomerRepresentativeAddress(ctx, customerID, representativeID, createAddress, optional)

Add customer representative address

Add an Address onto an existing Customer representative record

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to add the address onto | 
**representativeID** | **string**| representativeID of the Customer representative for whom to add the address | 
**createAddress** | [**CreateAddress**](CreateAddress.md)|  | 
 **optional** | ***AddCustomerRepresentativeAddressOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AddCustomerRepresentativeAddressOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**CustomerRepresentative**](CustomerRepresentative.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


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


## CreateCustomer

> Customer CreateCustomer(ctx, createCustomer, optional)

Create Customer

Create a Customer object from the given details of a human or business

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

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
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


## CreateCustomerAccount

> Account CreateCustomerAccount(ctx, customerID, createAccount, optional)

Create Customer Account

Create an account for the given customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to add an Account onto | 
**createAccount** | [**CreateAccount**](CreateAccount.md)|  | 
 **optional** | ***CreateCustomerAccountOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a CreateCustomerAccountOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DecryptAccountNumber

> TransitAccountNumber DecryptAccountNumber(ctx, customerID, accountID, optional)

Decrypt Account Number

Return the account number encrypted with a shared secret for application requests. This encryption key is different from the key used for persistence. 

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer the accountID belongs to | 
**accountID** | **string**| accountID of the Account to get decrypted account number | 
 **optional** | ***DecryptAccountNumberOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a DecryptAccountNumberOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**TransitAccountNumber**](TransitAccountNumber.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteAddress

> DeleteAddress(ctx, customerID, addressID)

Delete a customer's address

Deletes a customer's address

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
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


## DeleteCustomer

> DeleteCustomer(ctx, customerID, optional)

Delete Customer by ID

Remove a given Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to be deleted | 
 **optional** | ***DeleteCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a DeleteCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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


## DeleteCustomerAccount

> DeleteCustomerAccount(ctx, customerID, optional)

Delete Customer Account

Remove an account from the given Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to remove an Account | 
 **optional** | ***DeleteCustomerAccountOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a DeleteCustomerAccountOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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


## DeleteCustomerDocument

> DeleteCustomerDocument(ctx, customerID, documentID, optional)

Delete Customer Document

Remove Customer Document

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| ID of the customer that owns the document | 
**documentID** | **string**| ID of the document | 
 **optional** | ***DeleteCustomerDocumentOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a DeleteCustomerDocumentOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 

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


## DeleteCustomerRepresentative

> DeleteCustomerRepresentative(ctx, customerID, representativeID)

Delete a customer's representative

Deletes a customer's representative

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


## DeleteCustomerRepresentativeAddress

> DeleteCustomerRepresentativeAddress(ctx, customerID, representativeID, addressID)

Delete a customer representative's address

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


## GetCustomer

> Customer GetCustomer(ctx, customerID, optional)

Get Customer

Retrieve the Customer object and metadata for the customerID.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID that identifies this Customer | 
 **optional** | ***GetCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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


## GetCustomerAccountByID

> Account GetCustomerAccountByID(ctx, customerID, accountID, optional)

Get Customer Account

Retrieve an account by ID for the given customer.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to get an Account for | 
**accountID** | **string**| accountID of the Account | 
 **optional** | ***GetCustomerAccountByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerAccountByIDOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCustomerAccounts

> Customer AddCustomerAddress(ctx, customerID, createCustomerAddress, optional)

Add Customer Address

Add an Address onto an existing Customer record

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to add the address onto | 
**createCustomerAddress** | [**CreateCustomerAddress**](CreateCustomerAddress.md)|  | 
 **optional** | ***AddCustomerAddressOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a AddCustomerAddressOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
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


## CreateCustomer

> Customer CreateCustomer(ctx, createCustomer, optional)

Create Customer

Create a Customer object from the given details of a human or business

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

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
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


## DeleteCustomer

> DeleteCustomer(ctx, customerID, optional)

Delete Customer

Remove a given Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to be deleted | 
 **optional** | ***DeleteCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a DeleteCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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


## DeleteCustomerAddress

> DeleteCustomerAddress(ctx, customerID, addressID)

Delete Customer Address

Deletes a customer's address

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
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


## GetCustomer

> Customer GetCustomer(ctx, customerID, optional)

Get Customer

Retrieve the Customer object and metadata for the customerID.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID that identifies this Customer | 
 **optional** | ***GetCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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


## GetLatestOFACSearch

> OfacSearch GetLatestOFACSearch(ctx, customerID, optional)

Latest Customer OFAC search

Get the latest OFAC search for a Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to get latest OFAC search | 
 **optional** | ***GetLatestOFACSearchOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetLatestOFACSearchOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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

Ping Customers Service

Check the Customers service to check if running

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

Refresh Customer OFAC search

Refresh OFAC search for a given Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to refresh OFAC search | 
 **optional** | ***RefreshOFACSearchOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a RefreshOFACSearchOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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

Update Customer Metadata

Replace the metadata object for a customer. Metadata is a map of unique keys associated to values to act as foreign key relationships or arbitrary data associated to a Customer.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to add the metadata onto | 
**customerMetadata** | [**CustomerMetadata**](CustomerMetadata.md)|  | 
 **optional** | ***ReplaceCustomerMetadataOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a ReplaceCustomerMetadataOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
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


## SearchCustomers

> []Customer SearchCustomers(ctx, optional)

Search Customers

Search for customers using different filter parameters

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SearchCustomersOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a SearchCustomersOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **query** | **optional.String**| Optional parameter for searching by customer name | 
 **email** | **optional.String**| Optional parameter for searching by customer email | 
 **status** | **optional.String**| Optional parameter for searching by customer status | 
 **type_** | **optional.String**| Optional parameter for searching by customer type | 
 **skip** | **optional.String**| Optional parameter for searching for customers by skipping over an initial group | 
 **count** | **optional.String**| Optional parameter for searching by specifying the amount to return | 
 **customerIDs** | **optional.String**| Optional parameter for searching by customers&#39; IDs | 

### Return type

[**[]Customer**](Customer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateAccountStatus

> Account UpdateAccountStatus(ctx, customerID, accountID, updateAccountStatus)

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

[**Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateAddress

> UpdateAddress(ctx, customerID, addressID, updateAddress)

Update customer's address

Updates the specified customer address

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
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


## UpdateCustomer

> Customer UpdateCustomer(ctx, customerID, createCustomer, optional)

Update customer

Update a Customer object

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID that identifies this Customer | 
**createCustomer** | [**CreateCustomer**](CreateCustomer.md)|  | 
 **optional** | ***UpdateCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UpdateCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
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


## UpdateCustomerRepresentative

> UpdateCustomerRepresentative(ctx, customerID, representativeID, createCustomerRepresentative)

Update customer representative

Updates the specified customer representative

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| Customer ID | 
**representativeID** | **string**| Representative ID | 
**createCustomerRepresentative** | [**CreateCustomerRepresentative**](CreateCustomerRepresentative.md)|  | 

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


## UpdateCustomerRepresentativeAddress

> UpdateCustomerRepresentativeAddress(ctx, customerID, representativeID, addressID, updateAddress)

Update customer representative's address

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


## UpdateCustomerStatus

> Customer UpdateCustomerStatus(ctx, customerID, updateCustomerStatus, optional)

Update Customer Status

Update the status for a customer, which can only be updated by authenticated users with permissions.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to update the CustomerStatus | 
**updateCustomerStatus** | [**UpdateCustomerStatus**](UpdateCustomerStatus.md)|  | 
 **optional** | ***UpdateCustomerStatusOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UpdateCustomerStatusOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
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

