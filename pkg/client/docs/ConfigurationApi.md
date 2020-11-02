# \ConfigurationApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetOrganizationConfiguration**](ConfigurationApi.md#GetOrganizationConfiguration) | **Get** /configuration/customers | Get Organization Configuration
[**GetOrganizationLogo**](ConfigurationApi.md#GetOrganizationLogo) | **Get** /configuration/logo | Get Organization Logo
[**UpdateOrganizationConfiguration**](ConfigurationApi.md#UpdateOrganizationConfiguration) | **Put** /configuration/customers | Update Organization Configuration
[**UploadOrganizationLogo**](ConfigurationApi.md#UploadOrganizationLogo) | **Put** /configuration/logo | Update Organization Logo



## GetOrganizationConfiguration

> OrganizationConfiguration GetOrganizationConfiguration(ctx, optional)

Get Organization Configuration

Retrieve current configuration for the provided organization.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***GetOrganizationConfigurationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetOrganizationConfigurationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**OrganizationConfiguration**](OrganizationConfiguration.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetOrganizationLogo

> *os.File GetOrganizationLogo(ctx, xOrganization)

Get Organization Logo

Retrieve the organization's logo

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**xOrganization** | **string**| Value used to separate and identify models | 

### Return type

[***os.File**](*os.File.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: image/png, image/jpg, image/svg+xml, image/gif, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateOrganizationConfiguration

> OrganizationConfiguration UpdateOrganizationConfiguration(ctx, organizationConfiguration, optional)

Update Organization Configuration

Update the configuration for the provided organization.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**organizationConfiguration** | [**OrganizationConfiguration**](OrganizationConfiguration.md)|  | 
 **optional** | ***UpdateOrganizationConfigurationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UpdateOrganizationConfigurationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**OrganizationConfiguration**](OrganizationConfiguration.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadOrganizationLogo

> OrganizationConfiguration UploadOrganizationLogo(ctx, xOrganization, file)

Update Organization Logo

Upload an organization's logo, or update it if it already exists

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**xOrganization** | **string**| Value used to separate and identify models | 
**file** | ***os.File*****os.File**| Logo image file to be uploaded | 

### Return type

[**OrganizationConfiguration**](OrganizationConfiguration.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

