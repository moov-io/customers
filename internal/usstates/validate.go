// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package usstates

import "strings"

var (
	// These abbreviations are mappings of the 50 United States of America, various
	// territories and federal districts.
	//
	// https://en.wikipedia.org/wiki/List_of_states_and_territories_of_the_United_States
	abbreviations = map[string]bool{
		// States
		"AL": true, // Alabama
		"AK": true, // Alaska
		"AZ": true, // Arizona
		"AR": true, // Arkansas
		"CA": true, // California
		"CO": true, // Colorado
		"CT": true, // Connecticut
		"DE": true, // Delaware
		"FL": true, // Florida
		"GA": true, // Georgia
		"HI": true, // Hawaii
		"ID": true, // Idaho
		"IL": true, // Illinois
		"IN": true, // Indiana
		"IA": true, // Iowa
		"KS": true, // Kansas
		"KY": true, // Kentucky
		"LA": true, // Louisiana
		"ME": true, // Maine
		"MD": true, // Maryland
		"MA": true, // Massachusetts
		"MI": true, // Michigan
		"MN": true, // Minnesota
		"MS": true, // Mississippi
		"MO": true, // Missouri
		"MT": true, // Montana
		"NE": true, // Nebraska
		"NV": true, // Nevada
		"NH": true, // New Hampshire
		"NJ": true, // New Jersey
		"NM": true, // New Mexico
		"NY": true, // New York
		"NC": true, // North Carolina
		"ND": true, // North Dakota
		"OH": true, // Ohio
		"OK": true, // Oklahoma
		"OR": true, // Oregon
		"PA": true, // Pennsylvania
		"RI": true, // Rhode Island
		"SC": true, // South Carolina
		"SD": true, // South Dakota
		"TN": true, // Tennessee
		"TX": true, // Texas
		"UT": true, // Utah
		"VT": true, // Vermont
		"VA": true, // Virginia
		"WA": true, // Washington
		"WV": true, // West Virginia
		"WI": true, // Wisconsin
		"WY": true, // Wyoming

		// Federal districts
		"DC": true, // District of Columbia

		// Territories
		"AS": true, // American Samoa
		"GU": true, // Guam
		"MP": true, // Northern Mariana Islands
		"PR": true, // Puerto Rico
		"VI": true, // U.S. Virgin Islands
	}
)

// Valid returns true if the given abbreviation matches a United States state, territory, or federal district.
func Valid(state string) bool {
	if v, exists := abbreviations[strings.ToUpper(state)]; v && exists {
		return true
	}
	return false
}
