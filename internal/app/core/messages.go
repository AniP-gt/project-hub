package core

import "project-hub/internal/state"

type FetchProjectMsg struct {
	Project state.Project
	Items   []state.Item
}

type ItemUpdatedMsg struct {
	Index int
	Item  state.Item
}

type ErrMsg struct {
	Err error
}

func NewErrMsg(err error) ErrMsg {
	return ErrMsg{Err: err}
}

type DismissNotificationMsg struct {
	ID int
}

type EnterStatusSelectModeMsg struct{}

type DetailReadyMsg struct {
	Item state.Item
}

// ActionResultMsg is used to convey the result of a small side-effect action
// such as opening a browser URL or copying to clipboard. Handlers can turn
// this into a user-visible notification.
type ActionResultMsg struct{ Message string }
