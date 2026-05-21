package terminal

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Pelfox/gophkeeper/apps/client/internal/app"
	box "github.com/nyaosorg/go-box/v3"
)

// RequestVaultItemType prompts the user to select the kind of item to create.
func RequestVaultItemType() (app.PlaintextVaultItemType, error) {
	itemTypes := []app.PlaintextVaultItemType{
		app.PlaintextVaultItemTypePassword,
		app.PlaintextVaultItemTypeTextNote,
		app.PlaintextVaultItemTypeBinaryFile,
		app.PlaintextVaultItemTypeBankCard,
	}

	options := make([]string, len(itemTypes))
	for i, itemType := range itemTypes {
		options[i] = string(itemType)
	}

	fmt.Fprintln(os.Stdout, "Select item type:")
	selected, err := box.SelectIndex(options, false, os.Stdout)
	if err != nil {
		return "", fmt.Errorf("failed to select item type: %w", err)
	}

	if len(selected) == 0 {
		return "", fmt.Errorf("no item type selected")
	}
	fmt.Fprintln(os.Stdout)

	return itemTypes[selected[0]], nil
}

// RequestVaultItemPayload prompts the user for item-type-specific payload
// fields.
func RequestVaultItemPayload(itemType app.PlaintextVaultItemType) (any, error) {
	switch itemType {
	case app.PlaintextVaultItemTypePassword:
		payload, err := requestPasswordPayload()
		if err != nil {
			return nil, err
		}

		return payload, nil
	case app.PlaintextVaultItemTypeTextNote:
		payload, err := requestTextNotePayload()
		if err != nil {
			return nil, err
		}

		return payload, nil
	case app.PlaintextVaultItemTypeBinaryFile:
		payload, err := requestBinaryFilePayload()
		if err != nil {
			return nil, err
		}

		return payload, nil
	case app.PlaintextVaultItemTypeBankCard:
		payload, err := requestBankCardPayload()
		if err != nil {
			return nil, err
		}

		return payload, nil
	default:
		return nil, app.ErrInvalidVaultItemType
	}
}

func requestPasswordPayload() (*app.PasswordPayload, error) {
	password, err := RequestHiddenUserInput("Enter password to store: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	website, err := RequestUserInput("Enter website (optional): ")
	if err != nil {
		return nil, fmt.Errorf("failed to get website: %w", err)
	}

	payload := app.PasswordPayload{
		Password: password,
	}
	if website != "" {
		payload.Website = &website
	}

	return &payload, nil
}

func requestTextNotePayload() (*app.TextNotePayload, error) {
	text, err := RequestUserInput("Enter note text: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get note text: %w", err)
	}

	if text == "" {
		return nil, fmt.Errorf("note text cannot be empty")
	}

	return &app.TextNotePayload{Text: text}, nil
}

func requestBinaryFilePayload() (*app.BinaryFilePayload, error) {
	path, err := RequestUserInput("Enter file path: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get file path: %w", err)
	}

	if path == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	payload := app.BinaryFilePayload{
		Path: path,
		Data: data,
	}
	return &payload, nil
}

func requestBankCardPayload() (*app.BankCardPayload, error) {
	number, err := RequestUserInput("Enter bank card number: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get bank card number: %w", err)
	}

	number = strings.ReplaceAll(number, " ", "")
	if !isValidBankCardNumber(number) {
		return nil, fmt.Errorf("bank card number is invalid")
	}

	cvv, err := RequestHiddenUserInput("Enter CVV: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get CVV: %w", err)
	}

	if !isValidCVV(cvv) {
		return nil, fmt.Errorf("CVV is invalid")
	}

	expirationDate, err := RequestUserInput("Enter expiration date (MM/YY): ")
	if err != nil {
		return nil, fmt.Errorf("failed to get expiration date: %w", err)
	}

	if !isValidExpirationDate(expirationDate) {
		return nil, fmt.Errorf("expiration date is invalid")
	}

	holderName, err := RequestUserInput("Enter holder name: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get holder name: %w", err)
	}

	if holderName == "" {
		return nil, fmt.Errorf("holder name cannot be empty")
	}

	payload := app.BankCardPayload{
		Number:         number,
		CVV:            cvv,
		ExpirationDate: expirationDate,
		HolderName:     holderName,
	}

	return &payload, nil
}

func isValidBankCardNumber(number string) bool {
	if len(number) < 12 || len(number) > 19 {
		return false
	}

	sum := 0
	doubleDigit := false
	for i := len(number) - 1; i >= 0; i-- {
		if number[i] < '0' || number[i] > '9' {
			return false
		}

		digit := int(number[i] - '0')
		if doubleDigit {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		doubleDigit = !doubleDigit
	}

	return sum%10 == 0
}

func isValidCVV(cvv string) bool {
	matched, err := regexp.MatchString(`^\d{3,4}$`, cvv)
	return err == nil && matched
}

func isValidExpirationDate(expirationDate string) bool {
	matched, err := regexp.MatchString(`^(0[1-9]|1[0-2])/\d{2}$`, expirationDate)
	if err != nil || !matched {
		return false
	}

	month, err := strconv.Atoi(expirationDate[:2])
	if err != nil {
		return false
	}

	year, err := strconv.Atoi(expirationDate[3:])
	if err != nil {
		return false
	}

	now := time.Now()
	currentYear := now.Year() % 100
	currentMonth := int(now.Month())

	return year > currentYear || year == currentYear && month >= currentMonth
}
