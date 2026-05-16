package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// RequestUserInput reads user-submitted input from the stdin.
func RequestUserInput(prompt string) (string, error) {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	return strings.TrimSpace(value), nil
}

// RequestHiddenUserInput reads user-submitted input from the stdin, hiding
// what user is entering (e.g. password).
func RequestHiddenUserInput(prompt string) (string, error) {
	fmt.Print(prompt)

	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	value := string(bytes)
	return strings.TrimSpace(value), nil
}
