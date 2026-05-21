package cmd

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Pelfox/gophkeeper/apps/client/internal/api"
	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/spf13/cobra"
)

type listVaultsClientMock struct {
	api.ClientWithResponsesInterface

	response *api.GetVaultsResponse
	err      error
}

func (m *listVaultsClientMock) GetVaultsWithResponse(
	context.Context,
	...api.RequestEditorFn,
) (*api.GetVaultsResponse, error) {
	return m.response, m.err
}

// TestAuthCommandConstructors verifies auth command names.
func TestAuthCommandConstructors(t *testing.T) {
	application := &app.App{}

	if got := NewLoginCmd(application).Use; got != "login" {
		t.Fatalf("NewLoginCmd().Use = %q, want %q", got, "login")
	}
	if got := NewRegisterCmd(application).Use; got != "register" {
		t.Fatalf("NewRegisterCmd().Use = %q, want %q", got, "register")
	}
}

// TestVaultCommandTree verifies vault command names and expected subcommands.
func TestVaultCommandTree(t *testing.T) {
	command := NewRootVaultCmd(&app.App{})

	if got := command.Use; got != "vault" {
		t.Fatalf("NewRootVaultCmd().Use = %q, want %q", got, "vault")
	}

	assertSubcommand(t, command, "new")
	assertSubcommand(t, command, "list")
	assertSubcommand(t, command, "update")
	assertSubcommand(t, command, "delete")
	assertSubcommand(t, command, "item")
}

// TestVaultItemCommandTree verifies vault item command names and subcommands.
func TestVaultItemCommandTree(t *testing.T) {
	command := newVaultItemCmd(&app.App{})

	if got := command.Use; got != "item" {
		t.Fatalf("newVaultItemCmd().Use = %q, want %q", got, "item")
	}

	assertSubcommand(t, command, "create")
	assertSubcommand(t, command, "list")
	assertSubcommand(t, command, "view")
	assertSubcommand(t, command, "update")
	assertSubcommand(t, command, "delete")
}

// TestVaultListCommandEmpty verifies empty vault list command output.
func TestVaultListCommandEmpty(t *testing.T) {
	vaults := []protocol.ProtocolVault{}
	application := &app.App{
		Client: &listVaultsClientMock{
			response: &api.GetVaultsResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200:      &vaults,
			},
		},
	}

	output := captureStdout(t, func() {
		err := newVaultListCmd(application).RunE(&cobra.Command{}, nil)
		if err != nil {
			t.Fatalf("RunE returned error: %v", err)
		}
	})

	assertContains(t, output, "No vaults found.\n")
}

// TestSelectVaultEmpty verifies empty vault selection returns without error.
func TestSelectVaultEmpty(t *testing.T) {
	vaults := []protocol.ProtocolVault{}
	application := &app.App{
		Client: &listVaultsClientMock{
			response: &api.GetVaultsResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200:      &vaults,
			},
		},
	}

	output := captureStdout(t, func() {
		_, ok, err := selectVault(&cobra.Command{}, application, "Select vault:")
		if err != nil {
			t.Fatalf("selectVault returned error: %v", err)
		}
		if ok {
			t.Fatalf("expected no vault to be selected")
		}
	})

	assertContains(t, output, "No vaults found.\n")
}

// TestSelectVaultItemEmpty verifies empty item selection returns without error.
func TestSelectVaultItemEmpty(t *testing.T) {
	output := captureStdout(t, func() {
		_, ok, err := selectVaultItem(nil)
		if err != nil {
			t.Fatalf("selectVaultItem returned error: %v", err)
		}
		if ok {
			t.Fatalf("expected no vault item to be selected")
		}
	})

	assertContains(t, output, "No vault items found.\n")
}

func assertSubcommand(t *testing.T, command *cobra.Command, use string) {
	t.Helper()

	for _, subcommand := range command.Commands() {
		if subcommand.Use == use {
			return
		}
	}

	t.Fatalf("expected %q subcommand to exist", use)
}

func testHTTPResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     make(http.Header),
	}
}
