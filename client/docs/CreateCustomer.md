# CreateCustomer

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FirstName** | **string** | Given Name or First Name | 
**MiddleName** | **string** | Middle Name | [optional] 
**LastName** | **string** | Surname or Last Name | 
**NickName** | **string** | Name Customer is preferred to be called | [optional] 
**Suffix** | **string** | Customers name suffix. \&quot;Jr\&quot;, \&quot;PH.D.\&quot; | [optional] 
**Type** | [**CustomerType**](CustomerType.md) |  | [optional] 
**BirthDate** | [**time.Time**](time.Time.md) | Legal date of birth | 
**Email** | **string** | Primary email address of customer name@domain.com | 
**SSN** | **string** | Customer Social Security Number (SSN) | [optional] 
**Phones** | [**[]CreatePhone**](CreatePhone.md) |  | [optional] 
**Addresses** | [**[]CreateCustomerAddress**](CreateCustomerAddress.md) |  | 
**Metadata** | **map[string]string** | Map of unique keys associated to values to act as foreign key relationships or arbitrary data associated to a Customer. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


