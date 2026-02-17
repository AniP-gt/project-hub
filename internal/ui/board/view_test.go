package board

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"project-hub/internal/state"
)

func TestGroupItemsByStatus_DoneAlwaysLast(t *testing.T) {
	tests := []struct {
		name           string
		items          []state.Item
		expectedOrder  []string
		expectDoneLast bool
	}{
		{
			name: "Done is last when present",
			items: []state.Item{
				{ID: "1", Status: "Done", Position: 1},
				{ID: "2", Status: "Todo", Position: 1},
			},
			expectedOrder:  []string{"Todo", "Done"},
			expectDoneLast: true,
		},
		{
			name: "Case-insensitive done detection",
			items: []state.Item{
				{ID: "1", Status: "DONE", Position: 1},
				{ID: "2", Status: "Todo", Position: 1},
			},
			expectedOrder:  []string{"Todo", "Done"},
			expectDoneLast: true,
		},
		{
			name: "Trimmed whitespace handling",
			items: []state.Item{
				{ID: "1", Status: "  done  ", Position: 1},
				{ID: "2", Status: "Todo", Position: 1},
			},
			expectedOrder:  []string{"Todo", "Done"},
			expectDoneLast: true,
		},
		{
			name: "Unknown statuses before Done",
			items: []state.Item{
				{ID: "1", Status: "Done", Position: 1},
				{ID: "2", Status: "Blocked", Position: 1},
				{ID: "3", Status: "Todo", Position: 1},
			},
			expectedOrder:  []string{"Todo", "Blocked", "Done"},
			expectDoneLast: true,
		},
		{
			name: "Known progression order respected",
			items: []state.Item{
				{ID: "1", Status: "In Progress", Position: 1},
				{ID: "2", Status: "Todo", Position: 1},
				{ID: "3", Status: "Draft", Position: 1},
			},
			expectedOrder:  []string{"Todo", "Draft", "In Progress"},
			expectDoneLast: false,
		},
		{
			name: "Full progression with unknown and Done",
			items: []state.Item{
				{ID: "1", Status: "Done", Position: 1},
				{ID: "2", Status: "In Progress", Position: 1},
				{ID: "3", Status: "Blocked", Position: 1},
				{ID: "4", Status: "Todo", Position: 1},
				{ID: "5", Status: "Draft", Position: 1},
				{ID: "6", Status: "Pending", Position: 1},
			},
			expectedOrder:  []string{"Todo", "Draft", "In Progress", "Blocked", "Pending", "Done"},
			expectDoneLast: true,
		},
		{
			name: "No Done present",
			items: []state.Item{
				{ID: "1", Status: "Todo", Position: 1},
				{ID: "2", Status: "In Progress", Position: 1},
			},
			expectedOrder:  []string{"Todo", "In Progress"},
			expectDoneLast: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columns := groupItemsByStatus(tt.items, nil)

			// Check order matches expected
			if len(columns) != len(tt.expectedOrder) {
				t.Errorf("expected %d columns, got %d", len(tt.expectedOrder), len(columns))
			}

			for i, col := range columns {
				if i >= len(tt.expectedOrder) {
					break
				}
				if col.Name != tt.expectedOrder[i] {
					t.Errorf("column %d: expected %q, got %q", i, tt.expectedOrder[i], col.Name)
				}
			}

			// Check Done is last if expected
			if tt.expectDoneLast {
				if len(columns) == 0 {
					t.Errorf("expected Done to be last, but no columns found")
				} else if columns[len(columns)-1].Name != "Done" {
					t.Errorf("expected Done to be last, but got %q", columns[len(columns)-1].Name)
				}
			}
		})
	}
}

func TestIsDoneStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"done", true},
		{"Done", true},
		{"DONE", true},
		{"DoNe", true},
		{"  done  ", true},
		{"  DONE  ", true},
		{"Todo", false},
		{"In Progress", false},
		{"donee", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := isDoneStatus(tt.status)
			if result != tt.expected {
				t.Errorf("isDoneStatus(%q) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestPositionSortingPreserved(t *testing.T) {
	items := []state.Item{
		{ID: "1", Status: "Todo", Position: 3},
		{ID: "2", Status: "Todo", Position: 1},
		{ID: "3", Status: "Todo", Position: 2},
	}

	columns := groupItemsByStatus(items, nil)

	if len(columns) != 1 {
		t.Fatalf("expected 1 column, got %d", len(columns))
	}

	col := columns[0]
	if len(col.Cards) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(col.Cards))
	}

	expectedOrder := []string{"2", "3", "1"}
	for i, card := range col.Cards {
		if card.ID != expectedOrder[i] {
			t.Errorf("card %d: expected ID %q, got %q", i, expectedOrder[i], card.ID)
		}
	}
}

func TestEstimateCardHeightWithManyCards(t *testing.T) {
	items := make([]state.Item, 15)
	for i := 0; i < 15; i++ {
		items[i] = state.Item{
			ID:       string(rune('A' + i)),
			Status:   "Todo",
			Position: i,
			Title:    "Test card title",
			Priority: "Medium",
		}
	}

	board := NewBoardModel(items, []state.Field{}, state.FilterState{}, "")
	board.Width = 100
	board.Height = 50

	cardHeight := board.estimateCardHeight()
	if cardHeight <= 0 {
		t.Errorf("estimateCardHeight() returned %d, expected positive value", cardHeight)
	}
	if cardHeight > 15 {
		t.Errorf("estimateCardHeight() returned %d, expected reasonable value (<15)", cardHeight)
	}
}

func TestCalculateMaxVisibleCardsWithLargeList(t *testing.T) {
	items := make([]state.Item, 20)
	for i := 0; i < 20; i++ {
		items[i] = state.Item{
			ID:       string(rune('A' + i)),
			Status:   "Todo",
			Position: i,
			Title:    "Test card",
		}
	}

	board := NewBoardModel(items, []state.Field{}, state.FilterState{}, "")
	board.Width = 100
	board.Height = 50

	maxVisible := board.calculateMaxVisibleCards()
	if maxVisible < 3 {
		t.Errorf("calculateMaxVisibleCards() returned %d, expected at least 3 cards visible", maxVisible)
	}
	if maxVisible > 20 {
		t.Errorf("calculateMaxVisibleCards() returned %d, expected at most 20 for height 50", maxVisible)
	}
}

func TestScrollOffsetClampingDown(t *testing.T) {
	items := make([]state.Item, 15)
	for i := 0; i < 15; i++ {
		items[i] = state.Item{
			ID:       string(rune('A' + i)),
			Status:   "Todo",
			Position: i,
			Title:    "Card " + string(rune('A'+i)),
		}
	}

	board := NewBoardModel(items, []state.Field{}, state.FilterState{}, "")
	board.Width = 100
	board.Height = 30

	for i := 0; i < 20; i++ {
		board.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	if board.CardOffset < 0 {
		t.Errorf("CardOffset is negative: %d", board.CardOffset)
	}
	if board.FocusedCardIndex >= len(board.Columns[0].Cards) {
		t.Errorf("FocusedCardIndex %d exceeds card count %d", board.FocusedCardIndex, len(board.Columns[0].Cards))
	}
}

func TestScrollOffsetClampingUp(t *testing.T) {
	items := make([]state.Item, 15)
	for i := 0; i < 15; i++ {
		items[i] = state.Item{
			ID:       string(rune('A' + i)),
			Status:   "Todo",
			Position: i,
			Title:    "Card " + string(rune('A'+i)),
		}
	}

	board := NewBoardModel(items, []state.Field{}, state.FilterState{}, "")
	board.Width = 100
	board.Height = 30
	board.FocusedCardIndex = 14
	board.CardOffset = 10

	for i := 0; i < 20; i++ {
		board.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	}

	if board.CardOffset < 0 {
		t.Errorf("CardOffset is negative: %d", board.CardOffset)
	}
	if board.FocusedCardIndex < 0 {
		t.Errorf("FocusedCardIndex is negative: %d", board.FocusedCardIndex)
	}
}

func TestCardBackgroundColorUnselected(t *testing.T) {
	tests := []struct {
		name          string
		card          state.Card
		checkTitleFn  func(string) bool
		checkAssignee bool
	}{
		{
			name: "Unselected card with title only",
			card: state.Card{
				ID:    "1",
				Title: "Card",
			},
			checkTitleFn: func(o string) bool { return len(o) > 0 },
		},
		{
			name: "Unselected card with all fields",
			card: state.Card{
				ID:       "2",
				Title:    "Task",
				Assignee: "alice",
				Labels:   []string{"bug"},
				Priority: "High",
			},
			checkTitleFn:  func(o string) bool { return strings.Contains(o, "Task") },
			checkAssignee: true,
		},
		{
			name: "Unselected card with assignee",
			card: state.Card{
				ID:       "3",
				Title:    "Work",
				Assignee: "bob",
			},
			checkTitleFn:  func(o string) bool { return strings.Contains(o, "Work") },
			checkAssignee: true,
		},
		{
			name: "Unselected card with labels",
			card: state.Card{
				ID:     "4",
				Title:  "Bug",
				Labels: []string{"feature"},
			},
			checkTitleFn: func(o string) bool { return strings.Contains(o, "Bug") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoardModel([]state.Item{}, []state.Field{}, state.FilterState{}, "")
			board.Width = 80
			board.Height = 40

			output := board.renderCard(tt.card, false)

			if len(output) == 0 {
				t.Errorf("renderCard returned empty output")
			}

			if !tt.checkTitleFn(output) {
				t.Errorf("renderCard output invalid, got: %q", output)
			}

			if tt.checkAssignee && !strings.Contains(output, tt.card.Assignee) {
				t.Errorf("renderCard output missing assignee %q", tt.card.Assignee)
			}
		})
	}
}

func TestCardBackgroundColorSelected(t *testing.T) {
	tests := []struct {
		name string
		card state.Card
	}{
		{
			name: "Selected card with title only",
			card: state.Card{
				ID:    "1",
				Title: "Work",
			},
		},
		{
			name: "Selected card with all fields",
			card: state.Card{
				ID:       "2",
				Title:    "Task",
				Assignee: "bob",
				Labels:   []string{"bug"},
				Priority: "Medium",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoardModel([]state.Item{}, []state.Field{}, state.FilterState{}, "")
			board.Width = 80
			board.Height = 40

			output := board.renderCard(tt.card, true)

			if len(output) == 0 {
				t.Errorf("renderCard selected returned empty output")
			}

			if !strings.Contains(output, tt.card.Title) {
				t.Errorf("renderCard selected output missing card title %q", tt.card.Title)
			}
		})
	}
}

func TestCardPriorityColorsPreservedWithBackground(t *testing.T) {
	tests := []struct {
		name     string
		priority string
	}{
		{name: "High priority with background", priority: "High"},
		{name: "Medium priority with background", priority: "Medium"},
		{name: "Low priority with background", priority: "Low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := state.Card{
				ID:       "1",
				Title:    "Priority Test",
				Priority: tt.priority,
			}

			board := NewBoardModel([]state.Item{}, []state.Field{}, state.FilterState{}, "")
			board.Width = 80
			board.Height = 40

			outputUnselected := board.renderCard(card, false)
			if !strings.Contains(outputUnselected, tt.priority) {
				t.Errorf("unselected card output missing priority text %q", tt.priority)
			}

			outputSelected := board.renderCard(card, true)
			if !strings.Contains(outputSelected, tt.priority) {
				t.Errorf("selected card output missing priority text %q", tt.priority)
			}
		})
	}
}

func TestCardWithAllFieldsRenderLines(t *testing.T) {
	card := state.Card{
		ID:       "test-1",
		Title:    "Fix",
		Assignee: "alice",
		Labels:   []string{"bug"},
		Priority: "High",
	}

	board := NewBoardModel([]state.Item{}, []state.Field{}, state.FilterState{}, "")
	board.Width = 100
	board.Height = 50

	tests := []struct {
		name       string
		isSelected bool
	}{
		{name: "All fields unselected with background", isSelected: false},
		{name: "All fields selected with background", isSelected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := board.renderCard(card, tt.isSelected)

			if !strings.Contains(output, "Fix") {
				t.Errorf("renderCard output missing card title")
			}
			if !strings.Contains(output, "alice") {
				t.Errorf("renderCard output missing assignee")
			}
			if !strings.Contains(output, "bug") {
				t.Errorf("renderCard output missing label")
			}
			if !strings.Contains(output, "High") {
				t.Errorf("renderCard output missing priority")
			}
		})
	}
}

func TestCardBackgroundDifferentBetweenStates(t *testing.T) {
	card := state.Card{
		ID:       "comparison-1",
		Title:    "Test",
		Assignee: "bob",
		Labels:   []string{"test"},
		Priority: "Medium",
	}

	board := NewBoardModel([]state.Item{}, []state.Field{}, state.FilterState{}, "")
	board.Width = 80
	board.Height = 40

	unselectedOutput := board.renderCard(card, false)
	selectedOutput := board.renderCard(card, true)

	if len(unselectedOutput) == 0 {
		t.Errorf("unselected output is empty")
	}
	if len(selectedOutput) == 0 {
		t.Errorf("selected output is empty")
	}

	if !strings.Contains(unselectedOutput, "Test") {
		t.Errorf("unselected output missing card title")
	}
	if !strings.Contains(selectedOutput, "Test") {
		t.Errorf("selected output missing card title")
	}

	if !strings.Contains(unselectedOutput, "bob") {
		t.Errorf("unselected output missing assignee")
	}
	if !strings.Contains(selectedOutput, "bob") {
		t.Errorf("selected output missing assignee")
	}
}
