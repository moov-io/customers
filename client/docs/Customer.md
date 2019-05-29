# Customer

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier for the customer who owns the account | [optional] 
**FirstName** | **string** | Given Name or First Name | [optional] 
**MiddleName** | **string** | Middle Name | [optional] 
**LastName** | **string** | Surname or Last Name | [optional] 
**NickName** | **string** | Name Customer is preferred to be called | [optional] 
**Suffix** | **string** | Customers name suffix. \&quot;Jr\&quot;, \&quot;PH.D.\&quot; | [optional] 
**BirthDate** | [**time.Time**](time.Time.md) | Legal date of birth | [optional] 
**Status** | **string** | State of the customer | [optional] 
**Email** | **string** | Primary email address of customer name@domain.com | [optional] 
**Phones** | [**[]Phone**](Phone.md) |  | [optional] 
**Addresses** | [**[]Address**](Address.md) |  | [optional] 
**Metadata** | **map[string]string** |  | [optional] 
**CreatedAt** | [**time.Time**](time.Time.md) |  | [optional] 
**LastModified** | [**time.Time**](time.Time.md) | Last time the object was modified | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


