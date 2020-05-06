// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

// if customer.customerType == INDIVIDUAL then (firstName != "" && lastName != "")
// if customer.customerType == BUSINESS then (legalName != "")

// // ApprovedAt returns true only if the customerStatus is higher than ReviewRequired
// // and is at least the minimum status. It's used to ensure a specific customer is at least
// // KYC, OFAC, or CIP in applications.
// func (cs Status) ApprovedAt(minimum Status) bool {
// 	return ApprovedAt(cs, minimum)
// }

// // ApprovedAt returns true only if the customerStatus is higher than ReviewRequired
// // and is at least the minimum status. It's used to ensure a specific customer is at least
// // KYC, OFAC, or CIP in applications.
// func ApprovedAt(customerStatus Status, minimum Status) bool {
// 	if customerStatus <= ReviewRequired {
// 		return false // any status below ReveiewRequired is never approved
// 	}
// 	return customerStatus >= minimum
// }
