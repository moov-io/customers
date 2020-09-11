# InitAccountValidationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Strategy** | **string** | Validation strategy to use for the account.  micro-deposits: Initiate two small credits to the account along with a later balancing debit.  instant: Initiate instant account validation with specified vendor (e.g. Plaid, MX).  | 
**Vendor** | **string** |  | [optional] [default to VENDOR_MOOV]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


