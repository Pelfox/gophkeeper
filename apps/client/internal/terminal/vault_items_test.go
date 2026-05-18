package terminal

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
)

// TestRequestVaultItemPayloadInvalidType verifies invalid payload type
// selection is rejected before any interactive payload prompts are made.
func TestRequestVaultItemPayloadInvalidType(t *testing.T) {
	_, err := RequestVaultItemPayload(app.PlaintextVaultItemType("unknown"))
	if !errors.Is(err, app.ErrInvalidVaultItemType) {
		t.Fatalf("unexpected error: got %v want %v", err, app.ErrInvalidVaultItemType)
	}
}

// TestIsValidBankCardNumber verifies Luhn validation for card numbers.
func TestIsValidBankCardNumber(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{name: "valid visa", number: "4111111111111111", want: true},
		{name: "valid mastercard", number: "5555555555554444", want: true},
		{name: "invalid checksum", number: "4111111111111112", want: false},
		{name: "non digits", number: "411111111111111x", want: false},
		{name: "too short", number: "411111", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidBankCardNumber(tt.number); got != tt.want {
				t.Fatalf("isValidBankCardNumber() = %t, want %t", got, tt.want)
			}
		})
	}
}

// TestIsValidCVV verifies CVV validation for supported lengths.
func TestIsValidCVV(t *testing.T) {
	tests := []struct {
		name string
		cvv  string
		want bool
	}{
		{name: "three digits", cvv: "123", want: true},
		{name: "four digits", cvv: "1234", want: true},
		{name: "too short", cvv: "12", want: false},
		{name: "non digits", cvv: "12x", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidCVV(tt.cvv); got != tt.want {
				t.Fatalf("isValidCVV() = %t, want %t", got, tt.want)
			}
		})
	}
}

// TestIsValidExpirationDate verifies card expiration date format and freshness.
func TestIsValidExpirationDate(t *testing.T) {
	now := time.Now()
	future := now.AddDate(2, 0, 0)
	past := now.AddDate(-2, 0, 0)

	tests := []struct {
		name           string
		expirationDate string
		want           bool
	}{
		{
			name: "future date",
			expirationDate: fmt.Sprintf(
				"%02d/%02d",
				int(future.Month()),
				future.Year()%100,
			),
			want: true,
		},
		{
			name: "past date",
			expirationDate: fmt.Sprintf(
				"%02d/%02d",
				int(past.Month()),
				past.Year()%100,
			),
			want: false,
		},
		{name: "bad month", expirationDate: "13/30", want: false},
		{name: "bad format", expirationDate: "1/30", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidExpirationDate(tt.expirationDate); got != tt.want {
				t.Fatalf("isValidExpirationDate() = %t, want %t", got, tt.want)
			}
		})
	}
}
