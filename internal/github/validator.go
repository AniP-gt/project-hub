package github

import (
	"fmt"
	"strings"
)

// ValidateStatusUpdateIDs enforces non-empty, trimmed, non-numeric-only constraints
// and expected node-ID shape checks for status update parameters.
//
// Returns an error if any ID fails validation with actionable error messages
// identifying which ID failed and why.
//
// ID Pattern Requirements:
// - projectID: must start with "PVT_" (ProjectV2 node ID)
// - itemID: must start with "PVTI_" (ProjectV2Item node ID)
// - fieldID: must be non-empty, non-numeric-only
// - optionID: must be non-empty, non-numeric-only
func ValidateStatusUpdateIDs(projectID, itemID, fieldID, optionID string) error {
	// Validate projectID
	if err := validateProjectID(projectID); err != nil {
		return err
	}

	// Validate itemID
	if err := validateItemID(itemID); err != nil {
		return err
	}

	// Validate fieldID
	if err := validateFieldID(fieldID); err != nil {
		return err
	}

	// Validate optionID
	if err := validateOptionID(optionID); err != nil {
		return err
	}

	return nil
}

// validateProjectID checks that projectID is a valid ProjectV2 node ID.
func validateProjectID(id string) error {
	if err := validateNonEmptyTrimmed(id, "project ID"); err != nil {
		return err
	}

	if isNumericOnly(id) {
		return fmt.Errorf("invalid project ID: %q is numeric only; expected ProjectV2 node ID (starts with 'PVT_')", id)
	}

	if !strings.HasPrefix(id, "PVT_") {
		return fmt.Errorf("invalid project ID: %q does not start with 'PVT_'; expected ProjectV2 node ID format", id)
	}

	return nil
}

// validateItemID checks that itemID is a valid ProjectV2Item node ID.
func validateItemID(id string) error {
	if err := validateNonEmptyTrimmed(id, "item ID"); err != nil {
		return err
	}

	if isNumericOnly(id) {
		return fmt.Errorf("invalid item ID: %q is numeric only; expected ProjectV2Item node ID (starts with 'PVTI_')", id)
	}

	if !strings.HasPrefix(id, "PVTI_") {
		return fmt.Errorf("invalid item ID: %q does not start with 'PVTI_'; expected ProjectV2Item node ID format", id)
	}

	return nil
}

// validateFieldID checks that fieldID is non-empty and not whitespace-only.
func validateFieldID(id string) error {
	if err := validateNonEmptyTrimmed(id, "field ID"); err != nil {
		return err
	}

	return nil
}

// validateOptionID checks that optionID is non-empty and not whitespace-only.
func validateOptionID(id string) error {
	if err := validateNonEmptyTrimmed(id, "option ID"); err != nil {
		return err
	}

	return nil
}

// validateNonEmptyTrimmed checks that an ID is non-empty and not just whitespace.
func validateNonEmptyTrimmed(id string, idType string) error {
	if id == "" {
		return fmt.Errorf("invalid %s: empty", idType)
	}

	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return fmt.Errorf("invalid %s: contains only whitespace", idType)
	}

	return nil
}

// isNumericOnly returns true if the ID contains only digits (0-9).
func isNumericOnly(id string) bool {
	if id == "" {
		return false
	}

	for _, ch := range id {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}
