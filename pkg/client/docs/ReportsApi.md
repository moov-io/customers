# \ReportsApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetReportOfCustomerAccounts**](ReportsApi.md#GetReportOfCustomerAccounts) | **Get** /reports/accounts | Create Report of Accounts



## GetReportOfCustomerAccounts

> []ReportAccountResponse GetReportOfCustomerAccounts(ctx, optional)

Create Report of Accounts

Retrieves a list of customer and account information.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***GetReportOfCustomerAccountsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetReportOfCustomerAccountsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 
 **xOrganization** | **optional.String**| Value used to separate and identify models | 
 **accountIDs** | **optional.String**| A list of customer account IDs with a limit of 25 IDs. | 

### Return type

[**[]ReportAccountResponse**](ReportAccountResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

