package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	edtechotp "example.com/edtech-otp-cutover"
)

type server struct {
	gateway *edtechotp.Gateway
	login   edtechotp.LoginService
}

type sendRequest struct {
	Phone     string `json:"phone"`
	RequestID string `json:"request_id"`
}

type verifyRequest struct {
	Phone     string `json:"phone"`
	Code      string `json:"code"`
	RequestID string `json:"request_id"`
	CourseID  string `json:"course_id"`
	LearnerID string `json:"learner_id"`
	Delivered bool   `json:"delivered"`
	Deadline  string `json:"deadline"`
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	gateway := edtechotp.NewGateway(key)
	s := server{gateway: gateway, login: edtechotp.LoginService{Verifier: gateway, Now: time.Now}}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/code", s.sendCode)
	mux.HandleFunc("POST /login/verify", s.verifyCode)
	log.Println("course login listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s server) sendCode(w http.ResponseWriter, r *http.Request) {
	var input sendRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Phone == "" || input.RequestID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone and request_id are required"})
		return
	}
	if err := s.gateway.SendCode(r.Context(), input.Phone, input.RequestID); err != nil {
		writeGatewayError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "code_sent"})
}

func (s server) verifyCode(w http.ResponseWriter, r *http.Request) {
	var input verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	deadline, err := time.Parse(time.RFC3339, input.Deadline)
	if err != nil || input.Phone == "" || input.Code == "" || input.RequestID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone, code, request_id, and RFC3339 deadline are required"})
		return
	}
	decision, err := s.login.VerifyLearner(r.Context(), edtechotp.CourseDelivery{
		CourseID: input.CourseID, LearnerID: input.LearnerID,
		Delivered: input.Delivered, Deadline: deadline,
	}, input.Phone, input.Code, input.RequestID)
	if err != nil {
		writeGatewayError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, decision)
}

func writeGatewayError(w http.ResponseWriter, err error) {
	var apiErr *edtechotp.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Code, "message": apiErr.Message})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "delivery request failed"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
