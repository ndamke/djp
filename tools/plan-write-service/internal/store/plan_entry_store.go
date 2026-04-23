package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type PlanEntryStore struct {
	dataRoot string
}

type PlanEntryFile struct {
	Path     string
	Document *FrontmatterDocument
}

func NewPlanEntryStore(dataRoot string) *PlanEntryStore {
	return &PlanEntryStore{dataRoot: filepath.Clean(dataRoot)}
}

func (s *PlanEntryStore) DataRoot() string {
	return s.dataRoot
}

func (s *PlanEntryStore) LoadPlanEntry(entryID string) (*PlanEntryFile, error) {
	path := filepath.Join(s.dataRoot, "content", "plan-eintraege", entryID+".md")
	doc, err := LoadFrontmatterDocument(path)
	if err != nil {
		return nil, err
	}

	return &PlanEntryFile{Path: path, Document: doc}, nil
}

func (s *PlanEntryStore) SavePlanEntry(entry *PlanEntryFile) error {
	tmpPath := entry.Path + ".tmp"
	payload, err := entry.Document.Render()
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return err
	}

	return os.Rename(tmpPath, entry.Path)
}

func (s *PlanEntryStore) ListPlanEntriesByPlan(planID string) ([]*PlanEntryFile, error) {
	dir := filepath.Join(s.dataRoot, "content", "plan-eintraege")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	results := make([]*PlanEntryFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		doc, err := LoadFrontmatterDocument(path)
		if err != nil {
			return nil, err
		}

		if doc.GetString("type") != "plan_eintrag" || doc.GetString("jahresplan_id") != planID {
			continue
		}

		results = append(results, &PlanEntryFile{Path: path, Document: doc})
	}

	return results, nil
}

func (s *PlanEntryStore) AssertReferenceExists(bezugTyp string, bezugID string) error {
	section := ""
	switch bezugTyp {
	case "fach":
		section = "faecher"
	case "projekt":
		section = "projekte"
	default:
		return fmt.Errorf("unsupported bezug_typ: %s", bezugTyp)
	}

	path := filepath.Join(s.dataRoot, "content", section, bezugID+".md")
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func (s *PlanEntryStore) AssertPlanExists(planID string) error {
	path := filepath.Join(s.dataRoot, "content", "jahresplaene", planID+".md")
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func (s *PlanEntryStore) AssertLernsituationExists(lernsituationID string) error {
	path := filepath.Join(s.dataRoot, "content", "lernsituationen", lernsituationID+".md")
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func (s *PlanEntryStore) LoadLernsituationMeta(lernsituationID string) (*FrontmatterDocument, error) {
	path := filepath.Join(s.dataRoot, "content", "lernsituationen", lernsituationID+".md")
	return LoadFrontmatterDocument(path)
}

func (s *PlanEntryStore) NextPlanEntryID() (string, error) {
	dir := filepath.Join(s.dataRoot, "content", "plan-eintraege")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	maxID := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		if !strings.HasPrefix(name, "pe-") {
			continue
		}

		numberPart := strings.TrimPrefix(name, "pe-")
		number, err := strconv.Atoi(numberPart)
		if err != nil {
			continue
		}

		if number > maxID {
			maxID = number
		}
	}

	return fmt.Sprintf("pe-%04d", maxID+1), nil
}
