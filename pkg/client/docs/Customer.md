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
**BusinessName** | **string** | Business Name for business type customers | [optional] 
**DoingBusinessAs** | **string** | Doing Business As (DBA) name for business type customers | [optional] 
**BusinessType** | [**BusinessType**](BusinessType.md) |  | [optional] 
**EIN** | **string** | Employer Identification Number (EIN) for business type customers | [optional] 
**DUNS** | **string** | Dun &amp; Bradstreet D-U-N-S Number (D-U-N-S) for business type customers | [optional] 
**SICCode** | [**SicCode**](SICCode.md) |  | [optional] 
**NAICSCode** | [**NaicsCode**](NAICSCode.md) |  | [optional] 
**BirthDate** | **string** | Legal date of birth | [optional] 
**Status** | [**CustomerStatus**](CustomerStatus.md) |  | 
**Email** | **string** | Primary email address of customer name@domain.com | 
**Website** | **string** | Company Website for business type customers | [optional] 
**DateBusinessEstablished** | **string** | Date business was established for business type customers | [optional] 
**Phones** | [**[]Phone**](Phone.md) |  | [optional] 
**Addresses** | [**[]Address**](Address.md) |  | [optional] 
**Representatives** | [**[]CustomerRepresentative**](CustomerRepresentative.md) |  | [optional] 
**Metadata** | **map[string]string** | Map of unique keys associated to values to act as foreign key relationships or arbitrary data associated to a Customer. | [optional] 
**CreatedAt** | [**time.Time**](time.Time.md) |  | 
**LastModified** | [**time.Time**](time.Time.md) | Last time the object was modified | 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


