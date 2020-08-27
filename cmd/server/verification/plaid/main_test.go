package plaid

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/base"

	// is it ok to introduce testify?
	"github.com/stretchr/testify/require"
)

func TestInitiateAccountVerificationHandler(t *testing.T) {
	customerID := base.ID()

	v, err := Factory()
	require.NoError(t, err)

	router := mux.NewRouter()
	router.HandleFunc("/customers/{customerID}/verify", v.InitiateAccountVerification())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/verify", customerID), nil)

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	require.Contains(t, body, "link_token")
	require.Contains(t, body, "expiration")
}
