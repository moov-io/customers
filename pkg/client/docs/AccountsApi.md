# \AccountsApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCustomerAccount**](AccountsApi.md#CreateCustomerAccount) | **Post** /customers/{customerID}/accounts | Create Customer Account
[**DecryptAccountNumber**](AccountsApi.md#DecryptAccountNumber) | **Post** /customers/{customerID}/accounts/{accountID}/decrypt | Decrypt Account Number
[**DeleteCustomerAccount**](AccountsApi.md#DeleteCustomerAccount) | **Delete** /customers/{customerID}/accounts/{accountID} | Delete Customer Account
[**GetCustomerAccountByID**](AccountsApi.md#GetCustomerAccountByID) | **Get** /customers/{customerID}/accounts/{accountID} | Get Customer Account
[**GetCustomerAccounts**](AccountsApi.md#GetCustomerAccounts) | **Get** /customers/{customerID}/accounts | Get Customer Accounts
[**GetLatestAccountOFACSearch**](AccountsApi.md#GetLatestAccountOFACSearch) | **Get** /customers/{customerID}/accounts/{accountID}/ofac | Latest Account OFAC Search
[**RefreshAccountOFACSearch**](AccountsApi.md#RefreshAccountOFACSearch) | **Put** /customers/{customerID}/accounts/{accountID}/refresh/ofac | Refresh Account OFAC Search
[**UpdateAccountStatus**](AccountsApi.md#UpdateAccountStatus) | **Put** /customers/{customerID}/accounts/{accountID}/status | Update Account Status



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


## DeleteCustomerAccount

> DeleteCustomerAccount(ctx, customerID, accountID, optional)

Delete Customer Account

Remove an account from the given Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerID** | **string**| customerID of the Customer to remove an Account | 
**accountID** | **string**| accountID of the Account | 
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


## GetLatestAccountOFACSearch

> OfacSearch GetLatestAccountOFACSearch(ctx, customerID, accountID, optional)

Latest Account OFAC Search

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


## RefreshAccountOFACSearch

> OfacSearch RefreshAccountOFACSearch(ctx, customerID, accountID, optional)

Refresh Account OFAC Search

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

