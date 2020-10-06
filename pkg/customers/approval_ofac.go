// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	moovhttp "github.com/moov-io/base/http"
	watchmanClient "github.com/moov-io/watchman/client"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/route"
	"github.com/moov-io/customers/pkg/watchman"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
)

var (
	ofacMatchThreshold float32 = func() float32 {
		if v := os.Getenv("OFAC_MATCH_THRESHOLD"); v != "" {
			f, err := strconv.ParseFloat(v, 32)
			if err == nil && f > 0.00 {
				return float32(f)
			}
		}
		return 0.99 // default, 99%
	}()
)

type ofacSearchResult struct {
	EntityID  string    `json:"entityID"`
	SDNName   string    `json:"sdnName"`
	SDNType   string    `json:"sdnType"`
	Match     float32   `json:"match"`
	CreatedAt time.Time `json:"createdAt"`
}

type OFACSearcher struct {
	repo           CustomerRepository
	watchmanClient watchman.Client
}

func NewOFACSearcher(repo CustomerRepository, client watchman.Client) *OFACSearcher {
	return &OFACSearcher{
		repo:           repo,
		watchmanClient: client,
	}
}

// storeCustomerOFACSearch performs OFAC searches against the Customer's name and nickname if populated.
// The higher matching search result is stored in s.customerRepository for use later (in approvals)
func (s *OFACSearcher) storeCustomerOFACSearch(cust *client.Customer, requestID string) error {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	if cust == nil {
		return errors.New("nil Customer")
	}

	sdn, err := s.watchmanClient.Search(ctx, formatCustomerName(cust), requestID)
	if err != nil {
		return fmt.Errorf("OFACSearcher.storeCustomerOFACSearch: name search for customer=%s: %v", cust.CustomerID, err)
	}
	var nickSDN *watchmanClient.OfacSdn
	if cust.NickName != "" {
		nickSDN, err = s.watchmanClient.Search(ctx, cust.NickName, requestID)
		if err != nil {
			return fmt.Errorf("OFACSearcher.storeCustomerOFACSearch: nickname search for customer=%s: %v", cust.CustomerID, err)
		}
	}
	// Save the higher matching SDN (from name search or nick name)
	switch {
	case nickSDN != nil && nickSDN.Match > sdn.Match:
		err = s.repo.saveCustomerOFACSearch(cust.CustomerID, ofacSearchResult{
			EntityID:  nickSDN.EntityID,
			SDNName:   nickSDN.SdnName,
			SDNType:   nickSDN.SdnType,
			Match:     nickSDN.Match,
			CreatedAt: time.Now(),
		})
	case sdn != nil:
		err = s.repo.saveCustomerOFACSearch(cust.CustomerID, ofacSearchResult{
			EntityID:  sdn.EntityID,
			SDNName:   sdn.SdnName,
			SDNType:   sdn.SdnType,
			Match:     sdn.Match,
			CreatedAt: time.Now(),
		})
	}
	if err != nil {
		return fmt.Errorf("OFACSearcher.storeCustomerOFACSearch: saveCustomerOFACSearch customer=%s: %v", cust.CustomerID, err)
	}
	return nil
}

func AddOFACRoutes(logger log.Logger, r *mux.Router, repo CustomerRepository, ofac *OFACSearcher) {
	logger = logger.WithKeyValue("package", "ofac")

	r.Methods("GET").Path("/customers/{customerID}/ofac").HandlerFunc(getLatestCustomerOFACSearch(logger, repo))
	r.Methods("PUT").Path("/customers/{customerID}/refresh/ofac").HandlerFunc(refreshOFACSearch(logger, repo, ofac))
}

func getLatestCustomerOFACSearch(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		customerID := route.GetCustomerID(w, r)
		if customerID == "" {
			return
		}

		result, err := repo.getLatestCustomerOFACSearch(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func refreshOFACSearch(logger log.Logger, repo CustomerRepository, ofac *OFACSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		requestID := moovhttp.GetRequestID(r)
		customerID := route.GetCustomerID(w, r)
		if customerID == "" {
			return
		}

		cust, err := repo.getCustomer(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Log(fmt.Sprintf("running live OFAC search for customer=%s", customerID))

		if err := ofac.storeCustomerOFACSearch(cust, requestID); err != nil {
			logger.LogErrorF("error refreshing ofac search: %v", err)
			moovhttp.Problem(w, err)
			return
		}
		result, err := repo.getLatestCustomerOFACSearch(customerID)
		if err != nil {
			logger.LogErrorF("error getting latest ofac search: %v", err)
			moovhttp.Problem(w, err)
			return
		}
		if result.Match > ofacMatchThreshold {
			logger.LogErrorF("customer=%s matched against OFAC entity=%s with a score of %.2f - rejecting customer", cust.CustomerID, result.EntityID, result.Match)

			if err := repo.updateCustomerStatus(cust.CustomerID, client.REJECTED, "manual OFAC refresh"); err != nil {
				logger.LogErrorF("error updating customer=%s error=%v", cust.CustomerID, err)
				moovhttp.Problem(w, err)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}
