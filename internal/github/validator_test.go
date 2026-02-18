package github

import (
	"strings"
	"testing"
)

// TestValidateStatusUpdateIDs tests the preflight validator for status update IDs.
func TestValidateStatusUpdateIDs(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		itemID    string
		fieldID   string
		optionID  string
		wantErr   bool
		errMsg    string
	}{
		// Valid cases
		{
			name:      "valid all IDs",
			projectID: "PVT_abc123",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   false,
		},
		{
			name:      "valid with special characters",
			projectID: "PVT_1234567890abcdef",
			itemID:    "PVTI_fedcba0987654321",
			fieldID:   "field_status_id",
			optionID:  "option_id_1",
			wantErr:   false,
		},

		// Empty cases
		{
			name:      "empty projectID",
			projectID: "",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "empty",
		},
		{
			name:      "empty itemID",
			projectID: "PVT_abc123",
			itemID:    "",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "empty",
		},
		{
			name:      "empty fieldID",
			projectID: "PVT_abc123",
			itemID:    "PVTI_xyz789",
			fieldID:   "",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "empty",
		},
		{
			name:      "empty optionID",
			projectID: "PVT_abc123",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "",
			wantErr:   true,
			errMsg:    "empty",
		},

		// Whitespace cases
		{
			name:      "whitespace only projectID",
			projectID: "   ",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "whitespace",
		},
		{
			name:      "whitespace only itemID",
			projectID: "PVT_abc123",
			itemID:    "  \t\n  ",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "whitespace",
		},
		{
			name:      "whitespace only fieldID",
			projectID: "PVT_abc123",
			itemID:    "PVTI_xyz789",
			fieldID:   "\t \n",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "whitespace",
		},
		{
			name:      "whitespace only optionID",
			projectID: "PVT_abc123",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "   ",
			wantErr:   true,
			errMsg:    "whitespace",
		},

		// Numeric only cases
		{
			name:      "numeric only projectID",
			projectID: "9",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "numeric only",
		},
		{
			name:      "numeric only projectID (multi-digit)",
			projectID: "12345",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "numeric only",
		},
		{
			name:      "numeric only itemID",
			projectID: "PVT_abc123",
			itemID:    "999",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "numeric only",
		},
		{
			name:      "numeric only fieldID",
			projectID: "PVT_abc123",
			itemID:    "PVTI_xyz789",
			fieldID:   "42",
			optionID:  "OPT_456",
			wantErr:   false,
		},
		{
			name:      "numeric only optionID",
			projectID: "PVT_abc123",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "123",
			wantErr:   false,
		},

		// Invalid prefix cases
		{
			name:      "projectID missing PVT_ prefix",
			projectID: "INVALID_abc123",
			itemID:    "PVTI_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "does not start with 'PVT_'",
		},
		{
			name:      "itemID missing PVTI_ prefix",
			projectID: "PVT_abc123",
			itemID:    "PVT_xyz789",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "does not start with 'PVTI_'",
		},

		// Mixed failures (should report first error)
		{
			name:      "both projectID and itemID invalid",
			projectID: "123",
			itemID:    "456",
			fieldID:   "FIELD_001",
			optionID:  "OPT_456",
			wantErr:   true,
			errMsg:    "project ID",
		},

		// Edge cases for valid IDs
		{
			name:      "minimal valid IDs",
			projectID: "PVT_a",
			itemID:    "PVTI_b",
			fieldID:   "f",
			optionID:  "o",
			wantErr:   false,
		},
		{
			name:      "long valid IDs",
			projectID: "PVT_" + strings.Repeat("a", 100),
			itemID:    "PVTI_" + strings.Repeat("b", 100),
			fieldID:   "f" + strings.Repeat("i", 100),
			optionID:  "o" + strings.Repeat("p", 100),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatusUpdateIDs(tt.projectID, tt.itemID, tt.fieldID, tt.optionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatusUpdateIDs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateStatusUpdateIDs() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// TestValidateProjectID tests projectID validation specifically.
func TestValidateProjectID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid", "PVT_abc", false},
		{"empty", "", true},
		{"numeric", "9", true},
		{"whitespace", "  ", true},
		{"no prefix", "abc", true},
		{"wrong prefix", "PVTI_abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateItemID tests itemID validation specifically.
func TestValidateItemID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid", "PVTI_xyz", false},
		{"empty", "", true},
		{"numeric", "999", true},
		{"whitespace", "\t ", true},
		{"no prefix", "xyz", true},
		{"wrong prefix", "PVT_xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateItemID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateItemID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateFieldID tests fieldID validation specifically.
func TestValidateFieldID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid with letters", "FIELD_001", false},
		{"valid with mixed", "f1i2e3ld", false},
		{"empty", "", true},
		{"numeric only", "123", false},
		{"whitespace", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateOptionID tests optionID validation specifically.
func TestValidateOptionID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid with prefix", "OPT_456", false},
		{"valid single letter", "o", false},
		{"valid mixed", "opt123", false},
		{"empty", "", true},
		{"numeric only", "789", false},
		{"whitespace", "\n\t", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptionID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOptionID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIsNumericOnly tests the numeric-only check.
func TestIsNumericOnly(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"empty", "", false},
		{"single digit", "9", true},
		{"multiple digits", "123456", true},
		{"with letter", "12a3", false},
		{"single letter", "a", false},
		{"with space", "1 2", false},
		{"with underscore", "1_2", false},
		{"zero", "0", true},
		{"leading zero", "0123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNumericOnly(tt.id)
			if got != tt.want {
				t.Errorf("isNumericOnly(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}
