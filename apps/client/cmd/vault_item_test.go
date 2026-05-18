package cmd

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
	"github.com/google/uuid"
)

// TestTruncate verifies short, exact-limit and truncated text rendering.
func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		value string
		limit int
		want  string
	}{
		{name: "short", value: "secret", limit: 10, want: "secret"},
		{name: "exact", value: "secret", limit: 6, want: "secret"},
		{name: "truncated", value: "secret note", limit: 9, want: "secret..."},
		{name: "tiny limit", value: "secret", limit: 3, want: "sec"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncate(tt.value, tt.limit); got != tt.want {
				t.Fatalf("truncate() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestMaskedBankCardNumber verifies card masking keeps only the last digits.
func TestMaskedBankCardNumber(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   string
	}{
		{name: "full number", number: "4111111111111111", want: "**** 1111"},
		{name: "short number", number: "1234", want: "1234"},
		{name: "empty", number: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskedBankCardNumber(tt.number); got != tt.want {
				t.Fatalf("maskedBankCardNumber() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestVaultItemDetails verifies the compact details shown in vault item lists.
func TestVaultItemDetails(t *testing.T) {
	website := "https://example.com"
	tests := []struct {
		name string
		item app.PlaintextVaultItem
		want string
	}{
		{
			name: "password with website",
			item: app.PlaintextVaultItem{
				Payload: app.PasswordPayload{Website: &website},
			},
			want: "website: https://example.com",
		},
		{
			name: "password without website",
			item: app.PlaintextVaultItem{
				Payload: app.PasswordPayload{},
			},
			want: "website: -",
		},
		{
			name: "text note",
			item: app.PlaintextVaultItem{
				Payload: app.TextNotePayload{Text: strings.Repeat("a", 60)},
			},
			want: "text: " + strings.Repeat("a", 45) + "...",
		},
		{
			name: "binary file",
			item: app.PlaintextVaultItem{
				Payload: app.BinaryFilePayload{
					Path: "secret.bin",
					Data: []byte{1, 2, 3},
				},
			},
			want: "file: secret.bin (3 bytes)",
		},
		{
			name: "bank card",
			item: app.PlaintextVaultItem{
				Payload: app.BankCardPayload{
					Number:     "4111111111111111",
					HolderName: "User Name",
				},
			},
			want: "card: **** 1111, holder: User Name",
		},
		{name: "unknown", item: app.PlaintextVaultItem{}, want: "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vaultItemDetails(tt.item); got != tt.want {
				t.Fatalf("vaultItemDetails() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRenderVaultItemPassword verifies password item rendering.
func TestRenderVaultItemPassword(t *testing.T) {
	website := "https://example.com"
	item := testVaultItem(app.PlaintextVaultItem{
		Type: app.PlaintextVaultItemTypePassword,
		Name: "login",
		Payload: app.PasswordPayload{
			Password: "secret",
			Website:  &website,
		},
	})

	output := captureStdout(t, func() {
		if err := renderVaultItem(item); err != nil {
			t.Fatalf("renderVaultItem returned error: %v", err)
		}
	})

	assertContains(t, output, "Name: login\n")
	assertContains(t, output, "Type: password\n")
	assertContains(t, output, "Website: https://example.com\n")
	assertContains(t, output, "Password: secret\n")
}

// TestRenderVaultItemTextNote verifies text note item rendering.
func TestRenderVaultItemTextNote(t *testing.T) {
	item := testVaultItem(app.PlaintextVaultItem{
		Type:    app.PlaintextVaultItemTypeTextNote,
		Name:    "note",
		Payload: app.TextNotePayload{Text: "secret note"},
	})

	output := captureStdout(t, func() {
		if err := renderVaultItem(item); err != nil {
			t.Fatalf("renderVaultItem returned error: %v", err)
		}
	})

	assertContains(t, output, "Text:\n")
	assertContains(t, output, "secret note\n")
}

// TestRenderVaultItemBinaryFile verifies binary file rendering and temp save.
func TestRenderVaultItemBinaryFile(t *testing.T) {
	item := testVaultItem(app.PlaintextVaultItem{
		Type: app.PlaintextVaultItemTypeBinaryFile,
		Name: "file",
		Payload: app.BinaryFilePayload{
			Path: "secret.bin",
			Data: []byte("secret bytes"),
		},
	})

	output := captureStdout(t, func() {
		if err := renderVaultItem(item); err != nil {
			t.Fatalf("renderVaultItem returned error: %v", err)
		}
	})

	assertContains(t, output, "Path: secret.bin\n")
	assertContains(t, output, "Size: 12 bytes\n")
	assertContains(t, output, "Saved to: ")
}

// TestRenderVaultItemBankCard verifies bank card item rendering.
func TestRenderVaultItemBankCard(t *testing.T) {
	item := testVaultItem(app.PlaintextVaultItem{
		Type: app.PlaintextVaultItemTypeBankCard,
		Name: "card",
		Payload: app.BankCardPayload{
			Number:         "4111111111111111",
			CVV:            "123",
			ExpirationDate: "12/30",
			HolderName:     "User Name",
		},
	})

	output := captureStdout(t, func() {
		if err := renderVaultItem(item); err != nil {
			t.Fatalf("renderVaultItem returned error: %v", err)
		}
	})

	assertContains(t, output, "Number: 4111111111111111\n")
	assertContains(t, output, "CVV: 123\n")
	assertContains(t, output, "Expiration date: 12/30\n")
	assertContains(t, output, "Holder name: User Name\n")
}

// TestSaveBinaryPayloadToTemp verifies that binary payload data is saved.
func TestSaveBinaryPayloadToTemp(t *testing.T) {
	path, err := saveBinaryPayloadToTemp(app.BinaryFilePayload{
		Path: "secret.bin",
		Data: []byte("secret bytes"),
	})
	if err != nil {
		t.Fatalf("saveBinaryPayloadToTemp returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved payload: %v", err)
	}
	if string(data) != "secret bytes" {
		t.Fatalf("unexpected saved payload: got %q", data)
	}
}

func testVaultItem(plaintext app.PlaintextVaultItem) app.VaultItem {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	return app.VaultItem{
		ID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		VaultID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		CreatedAt: now,
		UpdatedAt: now,
		Plaintext: plaintext,
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close stdout writer: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}

	return string(output)
}

func assertContains(t *testing.T, value string, substring string) {
	t.Helper()

	if !strings.Contains(value, substring) {
		t.Fatalf("expected %q to contain %q", value, substring)
	}
}
