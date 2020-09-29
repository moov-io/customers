# \ConfigurationApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetConfiguration**](ConfigurationApi.md#GetConfiguration) | **Get** /configuration/customers | Get Configuration
[**UpdateConfiguration**](ConfigurationApi.md#UpdateConfiguration) | **Put** /configuration/customers | Update Configuration



## GetConfiguration

> NamespaceConfiguration GetConfiguration(ctx, optional)

Get Configuration

Retrieve current configuration for the provided organization.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***GetConfigurationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetConfigurationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**NamespaceConfiguration**](NamespaceConfiguration.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateConfiguration

> NamespaceConfiguration UpdateConfiguration(ctx, namespaceConfiguration, optional)

Update Configuration

Update the configuration for the provided organization.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**namespaceConfiguration** | [**NamespaceConfiguration**](NamespaceConfiguration.md)|  | 
 **optional** | ***UpdateConfigurationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UpdateConfigurationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**NamespaceConfiguration**](NamespaceConfiguration.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

