package validator

type StrategyKey struct {
	Strategy string
	Vendor   string
}

type Strategy interface {
	InitAccountValidation(userID, accountID, customerID string) (*VendorResponse, error)
}

type VendorResponse map[string]string

type testStrategy struct{}

func TestStrategy() Strategy {
	return &testStrategy{}
}

func (t *testStrategy) InitAccountValidation(userID, accountID, customerID string) (*VendorResponse, error) {
	return &VendorResponse{
		"test": "ok",
	}, nil
}
