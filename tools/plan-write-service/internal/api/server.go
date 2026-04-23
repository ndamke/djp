package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"plan-write-service/internal/service"
	"plan-write-service/internal/store"
	"plan-write-service/internal/validation"
)

type Config struct {
	DataRoot string
	Logger   *log.Logger
}

type Server struct {
	logger  *log.Logger
	service *service.PlanEntryService
	mux     *http.ServeMux
}

type healthResponse struct {
	OK      bool   `json:"ok"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type errorResponse struct {
	OK    bool             `json:"ok"`
	Error service.APIError `json:"error"`
}

type movePlanEntryRequest struct {
	Startwoche       int    `json:"startwoche"`
	Endwoche         int    `json:"endwoche"`
	BezugTyp         string `json:"bezug_typ"`
	BezugID          string `json:"bezug_id"`
	ExpectedRevision string `json:"expected_revision"`
}

type createPlanEntryRequest struct {
	JahresplanID    string `json:"jahresplan_id"`
	LernsituationID string `json:"lernsituation_id"`
	Startwoche      int    `json:"startwoche"`
	Endwoche        int    `json:"endwoche"`
	BezugTyp        string `json:"bezug_typ"`
	BezugID         string `json:"bezug_id"`
}

type removePlanEntryRequest struct {
	ExpectedRevision string `json:"expected_revision"`
}

func NewServer(cfg Config) http.Handler {
	entryStore := store.NewPlanEntryStore(cfg.DataRoot)
	svc := service.NewPlanEntryService(entryStore)

	s := &Server{
		logger:  cfg.Logger,
		service: svc,
		mux:     http.NewServeMux(),
	}

	s.routes()

	return s.withCORS(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/plan-eintraege", s.handlePlanEntriesCollection)
	s.mux.HandleFunc("/api/plan-eintraege/", s.handlePlanEntries)
}

func (s *Server) handlePlanEntriesCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		writeError(w, service.NewMethodNotAllowed("POST expected"))
		return
	}

	var req createPlanEntryRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, service.NewBadRequest("INVALID_JSON", "Request-JSON ist ungueltig", service.ErrorDetail{Field: "body", Issue: err.Error()}))
		return
	}

	createReq := service.CreateRequest{
		PlanID:          req.JahresplanID,
		LernsituationID: req.LernsituationID,
		Startwoche:      req.Startwoche,
		Endwoche:        req.Endwoche,
		BezugTyp:        req.BezugTyp,
		BezugID:         req.BezugID,
	}

	if err := validation.ValidateCreateRequest(createReq); err != nil {
		writeError(w, err)
		return
	}

	result, err := s.service.CreatePlanEntry(r.Context(), createReq)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, service.NewMethodNotAllowed("GET expected"))
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		OK:      true,
		Service: "plan-write-service",
		Version: "1.0.0",
	})
}

func (s *Server) handlePlanEntries(w http.ResponseWriter, r *http.Request) {
	if strings.Trim(r.URL.Path, "/") == "api/plan-eintraege" {
		s.handlePlanEntriesCollection(w, r)
		return
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		writeError(w, service.NewMethodNotAllowed("POST expected"))
		return
	}

	entryID, action, ok := parsePlanEntryPath(r.URL.Path)
	if !ok {
		writeError(w, service.NewNotFound("ENDPOINT_NOT_FOUND", "Endpoint nicht gefunden"))
		return
	}

	if err := validation.ValidatePlanEntryID(entryID); err != nil {
		writeError(w, err)
		return
	}

	if action == "move" {
		var req movePlanEntryRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeError(w, service.NewBadRequest("INVALID_JSON", "Request-JSON ist ungueltig", service.ErrorDetail{Field: "body", Issue: err.Error()}))
			return
		}

		moveReq := service.MoveRequest{
			Startwoche:       req.Startwoche,
			Endwoche:         req.Endwoche,
			BezugTyp:         req.BezugTyp,
			BezugID:          req.BezugID,
			ExpectedRevision: req.ExpectedRevision,
		}

		if err := validation.ValidateMoveRequest(moveReq); err != nil {
			writeError(w, err)
			return
		}

		result, err := s.service.MovePlanEntry(r.Context(), entryID, moveReq)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
		return
	}

	if action == "remove" {
		var req removePlanEntryRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil && err != io.EOF {
			writeError(w, service.NewBadRequest("INVALID_JSON", "Request-JSON ist ungueltig", service.ErrorDetail{Field: "body", Issue: err.Error()}))
			return
		}

		result, err := s.service.RemovePlanEntry(r.Context(), entryID, service.RemoveRequest{ExpectedRevision: req.ExpectedRevision})
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
		return
	}

	writeError(w, service.NewNotFound("ENDPOINT_NOT_FOUND", "Endpoint nicht gefunden"))
}

func parsePlanEntryPath(path string) (entryID string, action string, ok bool) {
	trimmed := strings.TrimPrefix(path, "/api/plan-eintraege/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	allowedOrigins := map[string]struct{}{
		"http://127.0.0.1:1313": {},
		"http://127.0.0.1:1314": {},
		"http://localhost:1313": {},
		"http://localhost:1314": {},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			}
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, apiErr *service.APIError) {
	writeJSON(w, apiErr.Status, errorResponse{OK: false, Error: *apiErr})
}
