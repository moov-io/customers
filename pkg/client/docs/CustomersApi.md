# \CustomersApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AcceptDisclaimer**](CustomersApi.md#AcceptDisclaimer) | **Post** /customers/{customerID}/disclaimers/{disclaimerID} | Accept customer disclaimer
[**AddCustomerAddress**](CustomersApi.md#AddCustomerAddress) | **Post** /customers/{customerID}/address | Add customer address
[**CompleteAccountValidation**](CustomersApi.md#CompleteAccountValidation) | **Put** /customers/{customerID}/accounts/{accountID}/validations | Complete Account Validation
[**CreateCustomer**](CustomersApi.md#CreateCustomer) | **Post** /customers | Create customer
[**CreateCustomerAccount**](CustomersApi.md#CreateCustomerAccount) | **Post** /customers/{customerID}/accounts | Create Customer Account
[**DecryptAccountNumber**](CustomersApi.md#DecryptAccountNumber) | **Post** /customers/{customerID}/accounts/{accountID}/decrypt | Decrypt Account Number
[**DeleteCustomer**](CustomersApi.md#DeleteCustomer) | **Delete** /customers/{customerID} | Delete Customer by ID
[**DeleteCustomerAccount**](CustomersApi.md#DeleteCustomerAccount) | **Delete** /customers/{customerID}/accounts | Delete Customer Account
[**DeleteCustomerAddress**](CustomersApi.md#DeleteCustomerAddress) | **Delete** /customers/{customerID}/addresses/{addressID} | Delete a customer&#39;s address
[**DeleteCustomerDocument**](CustomersApi.md#DeleteCustomerDocument) | **Delete** /customers/{customerID}/documents/{documentID} | Delete a customer&#39;s document
[**GetAccountValidation**](CustomersApi.md#GetAccountValidation) | **Get** /customers/{customerID}/accounts/{accountID}/validations/{validationID} | Get Account Validation
[**GetCustomer**](CustomersApi.md#GetCustomer) | **Get** /customers/{customerID} | Retrieve customer
[**GetCustomerAccountByID**](CustomersApi.md#GetCustomerAccountByID) | **Get** /customers/{customerID}/accounts/{accountID} | Get Customer Account by ID
[**GetCustomerAccounts**](CustomersApi.md#GetCustomerAccounts) | **Get** /customers/{customerID}/accounts | Get Customer Accounts
[**GetCustomerDisclaimers**](CustomersApi.md#GetCustomerDisclaimers) | **Get** /customers/{customerID}/disclaimers | Get customer disclaimers
[**GetCustomerDocumentContents**](CustomersApi.md#GetCustomerDocumentContents) | **Get** /customers/{customerID}/documents/{documentID} | Get customer document
[**GetCustomerDocuments**](CustomersApi.md#GetCustomerDocuments) | **Get** /customers/{customerID}/documents | Get customer documents
[**GetLatestAccountOFACSearch**](CustomersApi.md#GetLatestAccountOFACSearch) | **Get** /customers/{customerID}/accounts/{accountID}/ofac | Latest Account OFAC search
[**GetLatestOFACSearch**](CustomersApi.md#GetLatestOFACSearch) | **Get** /customers/{customerID}/ofac | Latest Customer OFAC search
[**InitAccountValidation**](CustomersApi.md#InitAccountValidation) | **Post** /customers/{customerID}/accounts/{accountID}/validations | Initiate Account Validation
[**Ping**](CustomersApi.md#Ping) | **Get** /ping | Ping Customers
[**RefreshAccountOFACSearch**](CustomersApi.md#RefreshAccountOFACSearch) | **Put** /customers/{customerID}/accounts/{accountID}/refresh/ofac | Refresh Account OFAC search
[**RefreshOFACSearch**](CustomersApi.md#RefreshOFACSearch) | **Put** /customers/{customerID}/refresh/ofac | Refresh Customer OFAC search
[**ReplaceCustomerMetadata**](CustomersApi.md#ReplaceCustomerMetadata) | **Put** /customers/{customerID}/metadata | Update customer metadata
[**SearchCustomers**](CustomersApi.md#SearchCustomers) | **Get** /customers | Get customers
[**UpdateAccountStatus**](CustomersApi.md#UpdateAccountStatus) | **Put** /customers/{customerID}/accounts/{accountID}/status | Update Account Status
[**UpdateCustomer**](CustomersApi.md#UpdateCustomer) | **Put** /customers/{customerID} | Update customer
[**UpdateCustomerAddress**](CustomersApi.md#UpdateCustomerAddress) | **Put** /customers/{customerID}/addresses/{addressID} | Update customer&#39;s address
[**UpdateCustomerStatus**](CustomersApi.md#UpdateCustomerStatus) | **Put** /customers/{customerID}/status | Update customer status
[**UploadCustomerDocument**](CustomersApi.md#UploadCustomerDocument) | **Post** /customers/{customerID}/documents | Upload document



## AcceptDisclaimer

> Disclaimer AcceptDisclaimer(ctx, customerID, disclaimerID, optional)

Accept customer disclaimer

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


## AddCustomerAddress

> Customer AddCustomerAddress(ctx, customerID, createCustomerAddress, optional)

Add customer address

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

Create customer

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


## DeleteCustomerAddress

> DeleteCustomerAddress(ctx, customerID, addressID)

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


## DeleteCustomerDocument

> DeleteCustomerDocument(ctx, customerID, documentID, optional)

Delete a customer's document

Remove a customer's document

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

Retrieve customer

Get the Customer object and metadata for the customerID.

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

Get Customer Account by ID

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

> []Account GetCustomerAccounts(ctx, customerID, optional)

Get Customer Accounts

Retrieve all accounts for the given customer.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to get Accounts for | 
 **optional** | ***GetCustomerAccountsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerAccountsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**[]Account**](Account.md)

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

Get customer disclaimers

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


## GetCustomerDocumentContents

> *os.File GetCustomerDocumentContents(ctx, customerID, documentID, optional)

Get customer document

Retrieve the referenced document

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to get a Document | 
**documentID** | **string**| documentID to identify a Document | 
 **optional** | ***GetCustomerDocumentContentsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerDocumentContentsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[***os.File**](*os.File.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/pdf, image/_*, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCustomerDocuments

> []Document GetCustomerDocuments(ctx, customerID, optional)

Get customer documents

Get documents for a customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to get all Documents | 
 **optional** | ***GetCustomerDocumentsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetCustomerDocumentsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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


## GetLatestAccountOFACSearch

> OfacSearch GetLatestAccountOFACSearch(ctx, customerID, accountID, optional)

Latest Account OFAC search

Get the latest OFAC search for an Account

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer | 
**accountID** | **string**| accountID of the Account to get latest OFAC search | 
 **optional** | ***GetLatestAccountOFACSearchOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetLatestAccountOFACSearchOpts struct


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


## Ping

> Ping(ctx, )

Ping Customers

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


## RefreshAccountOFACSearch

> OfacSearch RefreshAccountOFACSearch(ctx, customerID, accountID, optional)

Refresh Account OFAC search

Refresh OFAC search for a given Account

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to refresh OFAC search | 
**accountID** | **string**| accountID of the Account to get latest OFAC search | 
 **optional** | ***RefreshAccountOFACSearchOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a RefreshAccountOFACSearchOpts struct


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

Update customer metadata

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

Get customers

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
 **count** | **optional.String**| Optional parameter for searching for customers by specifying the amount to return | 

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


## UpdateCustomerAddress

> UpdateCustomerAddress(ctx, customerID, addressID, updateCustomerAddress)

Update customer's address

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

> Customer UpdateCustomerStatus(ctx, customerID, updateCustomerStatus, optional)

Update customer status

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


## UploadCustomerDocument

> Document UploadCustomerDocument(ctx, customerID, type_, file, optional)

Upload document

Upload a document for the given customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to add a document | 
**type_** | **string**| Document type (see Document type for values) | 
**file** | ***os.File*****os.File**| Document to be uploaded | 
 **optional** | ***UploadCustomerDocumentOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UploadCustomerDocumentOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

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

