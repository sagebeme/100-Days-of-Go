package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// bookingRequest is the body of POST /bookings.
type bookingRequest struct {
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
}

// New returns the API with every middleware in place, outermost first:
//
//	RequestID  so everything after it can read the ID
//	Logger     so it logs the final status, even the 500 that Recover writes
//	Recover    closest to the handlers, so it catches their panics
//
// Its routes are GET /health, which answers 200 "ok", and POST /bookings, which
// answers 201 and logs "booking created" with the request_id and the phone masked
// by MaskPhone.
func New(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("POST /bookings", func(w http.ResponseWriter, r *http.Request) {
		var req bookingRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"body must be valid JSON"}` + "\n"))
			return
		}
		logger.InfoContext(r.Context(), "booking created",
			slog.String("request_id", RequestIDFrom(r.Context())),
			slog.String("phone", MaskPhone(req.Phone)),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"booked"}` + "\n"))
	})
	// Your code here: wrap mux in RequestID, Logger and Recover with Chain, in
	// the order the doc comment gives.
	return mux
}
