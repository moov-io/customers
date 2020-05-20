# Customer

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CustomerID** | **string** | The unique identifier for the customer who owns the account | 
**FirstName** | **string** | Given Name or First Name | 
**MiddleName** | **string** | Middle Name | [optional] 
**LastName** | **string** | Surname or Last Name | 
**NickName** | **string** | Name Customer is preferred to be called | [optional] 
**Suffix** | **string** | Customers name suffix. \&quot;Jr\&quot;, \&quot;PH.D.\&quot; | [optional] 
**Type** | [**CustomerType**](CustomerType.md) |  | 
**BirthDate** | [**time.Time**](time.Time.md) | Legal date of birth | [optional] 
**Status** | [**CustomerStatus**](CustomerStatus.md) |  | 
**Email** | **string** | Primary email address of customer name@domain.com | 
**Phones** | [**[]Phone**](Phone.md) |  | [optional] 
**Addresses** | [**[]CustomerAddress**](CustomerAddress.md) |  | [optional] 
**Metadata** | **map[string]string** | Map of unique keys associated to values to act as foreign key relationships or arbitrary data associated to a Customer. | [optional] 
**CreatedAt** | [**time.Time**](time.Time.md) |  | 
**LastModified** | [**time.Time**](time.Time.md) | Last time the object was modified | 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


