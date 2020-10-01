# \ConfigurationApi

All URIs are relative to *http://localhost:8087*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetConfiguration**](ConfigurationApi.md#GetConfiguration) | **Get** /configuration/customers | Get Configuration
[**GetOrganizationLogo**](ConfigurationApi.md#GetOrganizationLogo) | **Get** /configuration/logo | Retreive an organization&#39;s logo
[**UpdateConfiguration**](ConfigurationApi.md#UpdateConfiguration) | **Put** /configuration/customers | Update Configuration
[**UploadOrganizationLogo**](ConfigurationApi.md#UploadOrganizationLogo) | **Put** /configuration/logo | Upload an organization&#39;s logo



## GetConfiguration

> OrganizationConfiguration GetConfiguration(ctx, optional)

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

Retreive an organization's logo

Retrieve a previously-uploaded logo image from an organization configuration

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


## UpdateConfiguration

> OrganizationConfiguration UpdateConfiguration(ctx, organizationConfiguration, optional)

Update Configuration

Update the configuration for the provided organization.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**organizationConfiguration** | [**OrganizationConfiguration**](OrganizationConfiguration.md)|  | 
 **optional** | ***UpdateConfigurationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UpdateConfigurationOpts struct


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

Upload an organization's logo

Update the organization's configuration to include a logo image file.

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

