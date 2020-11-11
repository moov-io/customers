# CreateCustomerIndividual

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
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
**Email** | **string** | Primary email address of customer name@domain.com | 
**SSN** | **string** | Customer Social Security Number (SSN) | [optional] 
**Website** | **string** | Company Website for business type customers | [optional] 
**DateBusinessEstablished** | **string** | Date business was established for business type customers | [optional] 
**Phones** | [**[]CreatePhone**](CreatePhone.md) |  | [optional] 
**Addresses** | [**[]CreateAddress**](CreateAddress.md) |  | [optional] 
**Representatives** | [**[]CreateRepresentative**](CreateRepresentative.md) |  | [optional] 
**Metadata** | **map[string]string** | Map of unique keys associated to values to act as foreign key relationships or arbitrary data associated to a Customer. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


