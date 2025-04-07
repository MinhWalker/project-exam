package validator

import (
	"regexp"
	"strings"
)

// EthereumValidator provides validation methods for Ethereum-related input
type EthereumValidator struct{}

// NewEthereumValidator creates a new EthereumValidator
func NewEthereumValidator() *EthereumValidator {
	return &EthereumValidator{}
}

// IsValidAddress checks if the provided string is a valid Ethereum address
func (v *EthereumValidator) IsValidAddress(address string) bool {
	// Ethereum addresses are 42 characters long (including '0x' prefix)
	// and contain only hexadecimal characters
	if len(address) != 42 {
		return false
	}

	// Check for 0x prefix
	if !strings.HasPrefix(address, "0x") {
		return false
	}

	// Check if the remainder is a valid hex string
	match, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address)
	return match
}

// FormatAddress ensures an Ethereum address is correctly formatted
func (v *EthereumValidator) FormatAddress(address string) string {
	// Remove any whitespace
	address = strings.TrimSpace(address)

	// Ensure lowercase except for checksum characters
	address = strings.ToLower(address)

	// Add 0x prefix if missing
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}

	return address
}
