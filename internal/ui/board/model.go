package board

import (
	"time"

	"project-hub/internal/state"
)

type BoardModel struct {
	Columns            []state.Column
	FocusedColumnIndex int
	FocusedCardIndex   int
	Width              int
	Height             int
	ColumnWidth        int
	VisibleColumns     int
	ColumnOffset       int
	CardOffset         int
	FieldVisibility    state.CardFieldVisibility
}

func NewBoardModel(items []state.Item, fields []state.Field, filter state.FilterState, focusedItemID string, fieldVisibility state.CardFieldVisibility) BoardModel {
	filteredItems := state.ApplyFilter(items, fields, filter, time.Now())
	columns := groupItemsByStatus(filteredItems, fields)

	focusedColumnIndex := 0
	focusedCardIndex := 0
	for colIdx, col := range columns {
		for cardIdx, card := range col.Cards {
			if card.ID == focusedItemID {
				focusedColumnIndex = colIdx
				focusedCardIndex = cardIdx
				break
			}
		}
	}

	return BoardModel{
		Columns:            columns,
		FocusedColumnIndex: focusedColumnIndex,
		FocusedCardIndex:   focusedCardIndex,
		ColumnOffset:       0,
		CardOffset:         0,
		FieldVisibility:    fieldVisibility,
	}
}
