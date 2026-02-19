package state

import "testing"

func TestParseFilterIterationShorthand(t *testing.T) {
	fs := ParseFilter("@current next previous")
	if len(fs.Iterations) != 3 {
		t.Fatalf("expected 3 iteration tokens, got %d", len(fs.Iterations))
	}
	if fs.Iterations[0] != "@current" {
		t.Errorf("expected first token to be @current, got %q", fs.Iterations[0])
	}
	if fs.Iterations[1] != "next" {
		t.Errorf("expected second token to be next, got %q", fs.Iterations[1])
	}
	if fs.Iterations[2] != "previous" {
		t.Errorf("expected third token to be previous, got %q", fs.Iterations[2])
	}
	if fs.Query != "" {
		t.Errorf("expected query to be empty, got %q", fs.Query)
	}
}

func TestParseFilterIterationList(t *testing.T) {
	fs := ParseFilter("iteration:@current,@previous,@next")
	if len(fs.Iterations) != 3 {
		t.Fatalf("expected 3 iteration tokens, got %d", len(fs.Iterations))
	}
	if fs.Iterations[0] != "@current" {
		t.Errorf("expected first token to be @current, got %q", fs.Iterations[0])
	}
	if fs.Iterations[1] != "@previous" {
		t.Errorf("expected second token to be @previous, got %q", fs.Iterations[1])
	}
	if fs.Iterations[2] != "@next" {
		t.Errorf("expected third token to be @next, got %q", fs.Iterations[2])
	}
}

func TestParseFilterGroupByToken(t *testing.T) {
	fs := ParseFilter("group:iteration label:bug")
	if fs.GroupBy != "iteration" {
		t.Fatalf("expected groupBy iteration, got %q", fs.GroupBy)
	}
	if len(fs.Labels) != 1 || fs.Labels[0] != "bug" {
		t.Fatalf("expected label bug, got %v", fs.Labels)
	}
	if fs.Query != "" {
		t.Errorf("expected query to be empty, got %q", fs.Query)
	}
}

func TestParseFilterFieldToken(t *testing.T) {
	fs := ParseFilter("Sprint:Q1 labels:bug")
	if len(fs.FieldFilters) != 1 {
		t.Fatalf("expected one field filter, got %d", len(fs.FieldFilters))
	}
	values := fs.FieldFilters["Sprint"]
	if len(values) != 1 || values[0] != "Q1" {
		t.Fatalf("expected Sprint=Q1, got %v", values)
	}
}

func TestParseFilterFieldTokenWithSpaces(t *testing.T) {
	fs := ParseFilter("\"Iteration Name\":\"Q1 Sprint\"")
	if len(fs.FieldFilters) != 1 {
		t.Fatalf("expected one field filter, got %d", len(fs.FieldFilters))
	}
	values := fs.FieldFilters["Iteration Name"]
	if len(values) != 1 || values[0] != "Q1 Sprint" {
		t.Fatalf("expected Iteration Name=Q1 Sprint, got %v", values)
	}
}
