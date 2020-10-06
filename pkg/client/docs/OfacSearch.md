# OfacSearch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EntityID** | **string** | SDN EntityID of the Entity | 
**Blocked** | **bool** | If the search resulted in a positive match against a sanctions list and should be blocked from making transfers or other operations. | 
**SdnName** | **string** | Name of the SDN entity | 
**SdnType** | **string** | SDN entity type | 
**Match** | **float32** | Percentage of similarity between the Customer name and this OFAC entity | 
**CreatedAt** | [**time.Time**](time.Time.md) |  | 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


