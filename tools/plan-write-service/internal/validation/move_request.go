package validation

import (
	"regexp"

	"plan-write-service/internal/service"
)

var planEntryIDPattern = regexp.MustCompile(`^pe-[a-z0-9-]+$`)

func ValidatePlanEntryID(entryID string) *service.APIError {
	if !planEntryIDPattern.MatchString(entryID) {
		return service.NewBadRequest("INVALID_PLAN_ENTRY_ID", "Planeintrag-ID ist ungueltig", service.ErrorDetail{Field: "id", Issue: entryID})
	}

	return nil
}

func ValidateMoveRequest(req service.MoveRequest) *service.APIError {
	if req.Startwoche < 1 || req.Startwoche > 40 {
		return service.NewBadRequest("VALIDATION_ERROR", "startwoche muss zwischen 1 und 40 liegen", service.ErrorDetail{Field: "startwoche", Issue: "out_of_range"})
	}

	if req.Endwoche < 1 || req.Endwoche > 40 {
		return service.NewBadRequest("VALIDATION_ERROR", "endwoche muss zwischen 1 und 40 liegen", service.ErrorDetail{Field: "endwoche", Issue: "out_of_range"})
	}

	if req.Endwoche < req.Startwoche {
		return service.NewBadRequest("VALIDATION_ERROR", "endwoche muss groesser oder gleich startwoche sein", service.ErrorDetail{Field: "endwoche", Issue: "before_startwoche"})
	}

	if req.BezugTyp == "" && req.BezugID == "" {
		return nil
	}

	if req.BezugTyp != "fach" && req.BezugTyp != "projekt" {
		return service.NewUnprocessable("INVALID_BEZUG_TYP", "bezug_typ muss 'fach' oder 'projekt' sein", service.ErrorDetail{Field: "bezug_typ", Issue: req.BezugTyp})
	}

	if req.BezugID == "" {
		return service.NewUnprocessable("MISSING_BEZUG_ID", "bezug_id ist erforderlich, wenn bezug_typ gesetzt ist", service.ErrorDetail{Field: "bezug_id", Issue: "missing"})
	}

	return nil
}

func ValidateCreateRequest(req service.CreateRequest) *service.APIError {
	if req.PlanID == "" {
		return service.NewBadRequest("VALIDATION_ERROR", "jahresplan_id ist erforderlich", service.ErrorDetail{Field: "jahresplan_id", Issue: "missing"})
	}

	if req.LernsituationID == "" {
		return service.NewBadRequest("VALIDATION_ERROR", "lernsituation_id ist erforderlich", service.ErrorDetail{Field: "lernsituation_id", Issue: "missing"})
	}

	moveValidation := ValidateMoveRequest(service.MoveRequest{
		Startwoche: req.Startwoche,
		Endwoche:   req.Endwoche,
		BezugTyp:   req.BezugTyp,
		BezugID:    req.BezugID,
	})
	if moveValidation != nil {
		return moveValidation
	}

	return nil
}
