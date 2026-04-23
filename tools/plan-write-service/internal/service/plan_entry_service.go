package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"plan-write-service/internal/store"
)

type ErrorDetail struct {
	Field string `json:"field,omitempty"`
	Issue string `json:"issue,omitempty"`
}

type APIError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
	Status  int           `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewBadRequest(code, message string, details ...ErrorDetail) *APIError {
	return &APIError{Code: code, Message: message, Details: details, Status: 400}
}

func NewNotFound(code, message string, details ...ErrorDetail) *APIError {
	return &APIError{Code: code, Message: message, Details: details, Status: 404}
}

func NewConflict(code, message string, details ...ErrorDetail) *APIError {
	return &APIError{Code: code, Message: message, Details: details, Status: 409}
}

func NewUnprocessable(code, message string, details ...ErrorDetail) *APIError {
	return &APIError{Code: code, Message: message, Details: details, Status: 422}
}

func NewMethodNotAllowed(message string) *APIError {
	return &APIError{Code: "METHOD_NOT_ALLOWED", Message: message, Status: 405}
}

func NewInternal(message string, err error) *APIError {
	return &APIError{Code: "INTERNAL_ERROR", Message: fmt.Sprintf("%s: %v", message, err), Status: 500}
}

type MoveRequest struct {
	Startwoche       int
	Endwoche         int
	BezugTyp         string
	BezugID          string
	ExpectedRevision string
}

type CreateRequest struct {
	PlanID          string
	LernsituationID string
	Startwoche      int
	Endwoche        int
	BezugTyp        string
	BezugID         string
}

type RemoveRequest struct {
	ExpectedRevision string
}

type MoveResult struct {
	OK       bool              `json:"ok"`
	ID       string            `json:"id"`
	Updated  MoveResultPayload `json:"updated"`
	Revision string            `json:"revision"`
	Warnings []string          `json:"warnings"`
}

type CreateResult struct {
	OK       bool              `json:"ok"`
	ID       string            `json:"id"`
	Updated  MoveResultPayload `json:"updated"`
	Revision string            `json:"revision"`
	Warnings []string          `json:"warnings"`
}

type RemoveResult struct {
	OK       bool     `json:"ok"`
	ID       string   `json:"id"`
	Revision string   `json:"revision"`
	Warnings []string `json:"warnings"`
}

type MoveResultPayload struct {
	Startwoche int    `json:"startwoche"`
	Endwoche   int    `json:"endwoche"`
	BezugTyp   string `json:"bezug_typ"`
	BezugID    string `json:"bezug_id"`
}

type PlanEntryService struct {
	store *store.PlanEntryStore
}

func NewPlanEntryService(store *store.PlanEntryStore) *PlanEntryService {
	return &PlanEntryService{store: store}
}

func (s *PlanEntryService) MovePlanEntry(_ context.Context, entryID string, req MoveRequest) (*MoveResult, *APIError) {
	entry, err := s.store.LoadPlanEntry(entryID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, NewNotFound("PLAN_ENTRY_NOT_FOUND", "Planeintrag wurde nicht gefunden", ErrorDetail{Field: "id", Issue: entryID})
		}

		return nil, NewInternal("Planeintrag konnte nicht gelesen werden", err)
	}

	if entry.Document.GetString("type") != "plan_eintrag" {
		return nil, NewUnprocessable("INVALID_TYPE", "Datei ist kein gueltiger Planeintrag", ErrorDetail{Field: "type", Issue: entry.Document.GetString("type")})
	}

	currentRevision := entry.Document.GetString("revision")
	if req.ExpectedRevision != "" && currentRevision != req.ExpectedRevision {
		return nil, NewConflict("REVISION_MISMATCH", "Planeintrag wurde zwischenzeitlich geaendert", ErrorDetail{Field: "expected_revision", Issue: currentRevision})
	}

	bezugTyp := entry.Document.GetString("bezug_typ")
	bezugID := entry.Document.GetString("bezug_id")
	if req.BezugTyp != "" {
		if req.BezugTyp != bezugTyp || req.BezugID != bezugID {
			return nil, NewUnprocessable("CROSS_ROW_MOVE_FORBIDDEN", "Verschieben zwischen Zeilen ist nicht erlaubt", ErrorDetail{Field: "bezug_id", Issue: "row_change_forbidden"})
		}
		bezugTyp = req.BezugTyp
		bezugID = req.BezugID
	}

	if err := s.store.AssertReferenceExists(bezugTyp, bezugID); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, NewNotFound("REFERENCE_NOT_FOUND", "Fach oder Projekt wurde nicht gefunden", ErrorDetail{Field: "bezug_id", Issue: bezugID})
		}

		return nil, NewInternal("Referenz konnte nicht geprueft werden", err)
	}

	planEntries, err := s.store.ListPlanEntriesByPlan(entry.Document.GetString("jahresplan_id"))
	if err != nil {
		return nil, NewInternal("Planeintraege konnten nicht geladen werden", err)
	}

	for _, candidate := range planEntries {
		if candidate.Document.GetString("id") == entryID {
			continue
		}

		if isInactive(candidate.Document) {
			continue
		}

		if candidate.Document.GetString("bezug_typ") != bezugTyp || candidate.Document.GetString("bezug_id") != bezugID {
			continue
		}

		start, err := getInt(candidate.Document, "startwoche")
		if err != nil {
			return nil, NewInternal("startwoche konnte nicht gelesen werden", err)
		}
		end, err := getInt(candidate.Document, "endwoche")
		if err != nil {
			return nil, NewInternal("endwoche konnte nicht gelesen werden", err)
		}

		if req.Startwoche <= end && req.Endwoche >= start {
			label := candidate.Document.GetString("anzeige_nummer")
			if text := candidate.Document.GetString("anzeige_kurztext"); text != "" {
				label = fmt.Sprintf("%s %s", label, text)
			}

			return nil, NewConflict(
				"PLAN_ENTRY_CONFLICT",
				fmt.Sprintf("Zielbereich ueberschneidet sich mit %s (W%d-%d)", label, start, end),
				ErrorDetail{Field: "startwoche", Issue: fmt.Sprintf("overlap:%s", candidate.Document.GetString("id"))},
			)
		}
	}

	entry.Document.SetInt("startwoche", req.Startwoche)
	entry.Document.SetInt("endwoche", req.Endwoche)
	entry.Document.SetString("bezug_typ", bezugTyp)
	entry.Document.SetString("bezug_id", bezugID)
	revision := time.Now().UTC().Format(time.RFC3339Nano)
	entry.Document.SetString("revision", revision)

	if err := s.store.SavePlanEntry(entry); err != nil {
		return nil, NewInternal("Planeintrag konnte nicht gespeichert werden", err)
	}

	return &MoveResult{
		OK: true,
		ID: entryID,
		Updated: MoveResultPayload{
			Startwoche: req.Startwoche,
			Endwoche:   req.Endwoche,
			BezugTyp:   bezugTyp,
			BezugID:    bezugID,
		},
		Revision: revision,
		Warnings: []string{},
	}, nil
}

func (s *PlanEntryService) CreatePlanEntry(_ context.Context, req CreateRequest) (*CreateResult, *APIError) {
	if err := s.store.AssertPlanExists(req.PlanID); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, NewNotFound("PLAN_NOT_FOUND", "Jahresplan wurde nicht gefunden", ErrorDetail{Field: "jahresplan_id", Issue: req.PlanID})
		}

		return nil, NewInternal("Jahresplan konnte nicht geprueft werden", err)
	}

	if err := s.store.AssertLernsituationExists(req.LernsituationID); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, NewNotFound("LERNSITUATION_NOT_FOUND", "Lernsituation wurde nicht gefunden", ErrorDetail{Field: "lernsituation_id", Issue: req.LernsituationID})
		}

		return nil, NewInternal("Lernsituation konnte nicht geprueft werden", err)
	}

	if err := s.store.AssertReferenceExists(req.BezugTyp, req.BezugID); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, NewNotFound("REFERENCE_NOT_FOUND", "Fach oder Projekt wurde nicht gefunden", ErrorDetail{Field: "bezug_id", Issue: req.BezugID})
		}

		return nil, NewInternal("Referenz konnte nicht geprueft werden", err)
	}

	planEntries, err := s.store.ListPlanEntriesByPlan(req.PlanID)
	if err != nil {
		return nil, NewInternal("Planeintraege konnten nicht geladen werden", err)
	}

	for _, candidate := range planEntries {
		if isInactive(candidate.Document) {
			continue
		}

		if candidate.Document.GetString("lernsituation_id") == req.LernsituationID &&
			candidate.Document.GetString("bezug_typ") == req.BezugTyp &&
			candidate.Document.GetString("bezug_id") == req.BezugID {
			return nil, NewConflict(
				"LS_ALREADY_ASSIGNED",
				"Diese Lernsituation ist in der ausgewaehlten Zeile bereits eingeplant",
				ErrorDetail{Field: "lernsituation_id", Issue: req.LernsituationID},
			)
		}

		if candidate.Document.GetString("bezug_typ") != req.BezugTyp || candidate.Document.GetString("bezug_id") != req.BezugID {
			continue
		}

		start, err := getInt(candidate.Document, "startwoche")
		if err != nil {
			return nil, NewInternal("startwoche konnte nicht gelesen werden", err)
		}
		end, err := getInt(candidate.Document, "endwoche")
		if err != nil {
			return nil, NewInternal("endwoche konnte nicht gelesen werden", err)
		}

		if req.Startwoche <= end && req.Endwoche >= start {
			label := candidate.Document.GetString("anzeige_nummer")
			if text := candidate.Document.GetString("anzeige_kurztext"); text != "" {
				label = fmt.Sprintf("%s %s", label, text)
			}

			return nil, NewConflict(
				"PLAN_ENTRY_CONFLICT",
				fmt.Sprintf("Zielbereich ueberschneidet sich mit %s (W%d-%d)", label, start, end),
				ErrorDetail{Field: "startwoche", Issue: fmt.Sprintf("overlap:%s", candidate.Document.GetString("id"))},
			)
		}
	}

	lsMeta, err := s.store.LoadLernsituationMeta(req.LernsituationID)
	if err != nil {
		return nil, NewInternal("Lernsituation konnte nicht gelesen werden", err)
	}

	nextID, err := s.store.NextPlanEntryID()
	if err != nil {
		return nil, NewInternal("Neue Planeintrag-ID konnte nicht erzeugt werden", err)
	}

	revision := time.Now().UTC().Format(time.RFC3339Nano)
	entryPath := filepath.Join(s.store.DataRoot(), "content", "plan-eintraege", nextID+".md")
	body := fmt.Sprintf("Planeintrag fuer %s in %s.\n", req.LernsituationID, req.BezugID)
	content := fmt.Sprintf("---\nid: %s\ntype: plan_eintrag\njahresplan_id: %s\nlernsituation_id: %s\nbezug_typ: %s\nbezug_id: %s\nstartwoche: %d\nendwoche: %d\nanzeige_nummer: %q\nanzeige_kurztext: %q\nbuild:\n    list: true\n    render: false\naktiv: true\ndraft: false\nrevision: %q\n---\n\n%s",
		nextID,
		req.PlanID,
		req.LernsituationID,
		req.BezugTyp,
		req.BezugID,
		req.Startwoche,
		req.Endwoche,
		defaultString(lsMeta.GetString("nummer"), "LS"),
		defaultString(lsMeta.GetString("kurztext"), strings.TrimSpace(lsMeta.GetString("title"))),
		revision,
		body,
	)

	tmpPath := entryPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0o644); err != nil {
		return nil, NewInternal("Neuer Planeintrag konnte nicht gespeichert werden", err)
	}

	if err := os.Rename(tmpPath, entryPath); err != nil {
		return nil, NewInternal("Neuer Planeintrag konnte nicht gespeichert werden", err)
	}

	return &CreateResult{
		OK: true,
		ID: nextID,
		Updated: MoveResultPayload{
			Startwoche: req.Startwoche,
			Endwoche:   req.Endwoche,
			BezugTyp:   req.BezugTyp,
			BezugID:    req.BezugID,
		},
		Revision: revision,
		Warnings: []string{},
	}, nil
}

func (s *PlanEntryService) RemovePlanEntry(_ context.Context, entryID string, req RemoveRequest) (*RemoveResult, *APIError) {
	entry, err := s.store.LoadPlanEntry(entryID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, NewNotFound("PLAN_ENTRY_NOT_FOUND", "Planeintrag wurde nicht gefunden", ErrorDetail{Field: "id", Issue: entryID})
		}

		return nil, NewInternal("Planeintrag konnte nicht gelesen werden", err)
	}

	currentRevision := entry.Document.GetString("revision")
	if req.ExpectedRevision != "" && currentRevision != req.ExpectedRevision {
		return nil, NewConflict("REVISION_MISMATCH", "Planeintrag wurde zwischenzeitlich geaendert", ErrorDetail{Field: "expected_revision", Issue: currentRevision})
	}

	revision := time.Now().UTC().Format(time.RFC3339Nano)
	entry.Document.SetString("aktiv", "false")
	entry.Document.SetString("draft", "false")
	entry.Document.SetString("revision", revision)

	if err := s.store.SavePlanEntry(entry); err != nil {
		return nil, NewInternal("Planeintrag konnte nicht entfernt werden", err)
	}

	return &RemoveResult{OK: true, ID: entryID, Revision: revision, Warnings: []string{}}, nil
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func isTruthy(value string) bool {
	v := strings.TrimSpace(strings.ToLower(value))
	return v == "true" || v == "1" || v == "yes"
}

func isInactive(doc *store.FrontmatterDocument) bool {
	if isTruthy(doc.GetString("draft")) {
		return true
	}

	aktiv := strings.TrimSpace(strings.ToLower(doc.GetString("aktiv")))
	return aktiv == "false" || aktiv == "0" || aktiv == "no"
}

func getInt(doc *store.FrontmatterDocument, key string) (int, error) {
	value := doc.GetString(key)
	if value == "" {
		return 0, fmt.Errorf("missing value for %s", key)
	}

	return strconv.Atoi(value)
}
