package core

import (
	"errors"
	"os/exec"
	"testing"
)

func TestOpenBrowserCmd_NoOpener(t *testing.T) {
	origLook := lookPath
	defer func() { lookPath = origLook }()
	lookPath = func(file string) (string, error) { return "", errors.New("not found") }

	cmd := OpenBrowserCmd("https://example.com")
	msg := cmd()
	if _, ok := msg.(ActionResultMsg); !ok {
		t.Fatalf("expected ActionResultMsg when no opener found, got %T", msg)
	}
}

func TestCopyToClipboardCmd_NoTool(t *testing.T) {
	origLook := lookPath
	defer func() { lookPath = origLook }()
	lookPath = func(file string) (string, error) { return "", errors.New("not found") }

	cmd := CopyToClipboardCmd("https://example.com")
	msg := cmd()
	if _, ok := msg.(ActionResultMsg); !ok {
		t.Fatalf("expected ActionResultMsg when no clipboard tool found, got %T", msg)
	}
}

func TestOpenBrowserCmd_Present(t *testing.T) {
	origLook := lookPath
	defer func() { lookPath = origLook }()
	lookPath = func(file string) (string, error) { return "/usr/bin/fake", nil }
	origExec := execCommand
	defer func() { execCommand = origExec }()
	execCommand = func(name string, arg ...string) *exec.Cmd { return exec.Command("true") }

	cmd := OpenBrowserCmd("https://example.com")
	_ = cmd()
}
