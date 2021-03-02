package route

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/moov-io/base/log"
)

// To avoid changes into base/http.Wrap and its dependencies (Fed, Watchman,
// ...) base/http.Wrap code is duplicated here. When we rewrite Customers
// routing based on templater we will get rid of this code. Agree? Thoughts?

// ResponseWriter implements Go's standard library http.ResponseWriter to complete HTTP requests
type ResponseWriter struct {
	http.ResponseWriter

	start   time.Time
	request *http.Request
	metric  metrics.Histogram

	headersWritten bool // set on WriteHeader

	log log.Logger
}

// WriteHeader sends an HTTP response header with the provided status code, records response duration,
// and optionally records the HTTP metadata in a go-kit log.Logger
func (w *ResponseWriter) WriteHeader(code int) {
	if w == nil || w.headersWritten {
		return
	}
	w.headersWritten = true

	// Headers
	SetAccessControlAllowHeaders(w, w.request.Header.Get("Origin"))
	defer w.ResponseWriter.WriteHeader(code)

	// Record route timing
	diff := time.Since(w.start)
	if w.metric != nil {
		w.metric.Observe(diff.Seconds())
	}

	// Skip Go's content sniff here to speed up response timing for client
	if w.ResponseWriter.Header().Get("Content-Type") == "" {
		w.ResponseWriter.Header().Set("Content-Type", "text/plain")
		w.ResponseWriter.Header().Set("X-Content-Type-Options", "nosniff")
	}

	if requestID := GetRequestID(w.request); requestID != "" && w.log != nil {
		w.log.With(log.Fields{
			"method":    log.String(w.request.Method),
			"path":      log.String(w.request.URL.Path),
			"status":    log.String(strconv.Itoa(code)),
			"duration":  log.String(diff.String()),
			"requestID": log.String(requestID),
		})
	}
}

// Wrap returns a ResponseWriter usable by applications. No parts of the Request are inspected or ResponseWriter modified.
func Wrap(logger log.Logger, m metrics.Histogram, w http.ResponseWriter, r *http.Request) *ResponseWriter {
	now := time.Now()
	return &ResponseWriter{
		ResponseWriter: w,
		start:          now,
		request:        r,
		metric:         m,
		log:            logger,
	}
}

// SetAccessControlAllowHeaders writes Access-Control-Allow-* headers to a response to allow
// for further CORS-allowed requests.
func SetAccessControlAllowHeaders(w http.ResponseWriter, origin string) {
	// Access-Control-Allow-Origin can't be '*' with requests that send credentials.
	// Instead, we need to explicitly set the domain (from request's Origin header)
	//
	// Allow requests from anyone's localhost and only from secure pages.
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "https://") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Cookie,X-User-Id,X-Request-Id,Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

// GetRequestID returns the Moov header value for request IDs
func GetRequestID(r *http.Request) string {
	return r.Header.Get("X-Request-Id")
}
