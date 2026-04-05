package handlers

import (
	"compress/gzip"
	"encoding/json"
	"expvar"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"golang.org/x/time/rate"
)


func (h *Handler) RateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var mu sync.Mutex
	var clients = make(map[string]*client)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.Config.Limiter.Enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				h.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			_, found := clients[ip]
			if !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(h.Config.Limiter.RPS), h.Config.Limiter.Burst)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				h.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

//* ----------------------------------------- Metrics Middlewares ----------------------------------------- *// 
// expvar counters and maps
var (
	totalRequestsReceived           = expvar.NewInt("total_requests_received")
	totalResponsesSent              = expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_microseconds")
	totalInFlightRequests           = expvar.NewInt("total_in_flight_requests")
	responsesByStatus               = expvar.NewMap("responses_by_status")
	responsesByMethod               = expvar.NewMap("responses_by_method")
)

// metricsResponseWriter wraps http.ResponseWriter to capture status code and bytes written
type metricsResponseWriter struct {
    http.ResponseWriter   // the original http.ResponseWriter
    statusCode int         // this will contain the status code we need
    headerWritten bool    // has the response headers already been written?
}

func (mw *metricsResponseWriter) Header() http.Header {
    return mw.ResponseWriter.Header()
}


// WriteHeader captures the status code
func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.ResponseWriter.WriteHeader(statusCode)
	if !mw.headerWritten {
        mw.statusCode = statusCode
        mw.headerWritten = true
    }
}

// Write captures the number of bytes written
func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
    mw.headerWritten = true
    return mw.ResponseWriter.Write(b)
}


func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
    return mw.ResponseWriter
}

// metricsEndpointHandler returns curated metrics without exposing default memstats
func (h *Handler) Metrics (next http.Handler) http.Handler {	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequestsReceived.Add(1)
		totalInFlightRequests.Add(1)

		// Wrap ResponseWriter to capture status code
		mrw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(mrw, r)

		// After response
		duration := time.Since(start)
		totalResponsesSent.Add(1)
		totalInFlightRequests.Add(-1)

		responsesByMethod.Add(r.Method, 1)
		responsesByStatus.Add(strconv.Itoa(mrw.statusCode), 1)
		totalProcessingTimeMicroseconds.Add(duration.Microseconds())
	})
}

func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Build a filtered snapshot
	snapshot := map[string]any{}
	// Pull selected expvar metrics if present
	expvar.Do(func(kv expvar.KeyValue) {
		switch kv.Key {
		case "version", "env", "goroutines", "database",
			"total_requests_received", "total_responses_sent",
			"total_in_flight_requests", "total_processing_time_microseconds",
			"responses_by_status", "responses_by_method":
			snapshot[kv.Key] = parseExpvarValue(kv.Value)
		}
	})

	// Write JSON using existing helper
	_ = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"metrics": snapshot}, nil)
}

// parseExpvarValue converts expvar.Var to a plain Go value for JSON encoding
func parseExpvarValue(v expvar.Var) any {
	// expvar values implement String() with JSON; attempt to decode that
	type jsonMarshaler interface{ String() string }
	if jm, ok := v.(jsonMarshaler); ok {
		s := jm.String()
		// Best-effort JSON decode
		var out any
		if err := json.Unmarshal([]byte(s), &out); err == nil {
			return out
		}
		return s
	}
	return v
}


// LoggingMiddleware logs each HTTP request in JSON format
func (h *Handler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code and bytes written
		wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Serve the request
		next.ServeHTTP(wrapped, r)

		// Log the request details in JSON format using slog
		h.Logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"status_code", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// gzipResponseWriter wraps http.ResponseWriter and compresses the response body
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.gzipWriter.Write(b)
}

func (grw *gzipResponseWriter) Close() error {
	return grw.gzipWriter.Close()
}

// GzipMiddleware compresses response bodies with gzip when client sends Accept-Encoding: gzip
func (h *Handler) GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Create a gzip writer
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			// Wrap the response writer
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length") // Content-Length will change with compression

			grw := &gzipResponseWriter{
				ResponseWriter: w,
				gzipWriter:     gzipWriter,
			}

			next.ServeHTTP(grw, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (h *Handler) EnableCORS(next http.Handler) http.Handler {                             
   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		// check the request origin to see if it's in the trusted list
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")

		// Once we have a origin from the request header we need to check
		if origin != "" {
			for i := range h.Config.CORS.TrustedOrigins {
				if origin == h.Config.CORS.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					if r.Method == http.MethodOptions && 
						r.Header.Get("Access-Control-Request-Method") != "" {
							w.Header().Set("Access-Control-Allow-Methods",
											"OPTIONS, PUT, PATCH, DELETE")
							w.Header().Set("Access-Control-Allow-Headers",
											"Authorization, Content-Type")
						
							// we need to send a 200 OK status. Also since there
							// is no need to continue the middleware chain we
							// we leave 
							w.WriteHeader(http.StatusOK)
							return
						}
					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})

}

