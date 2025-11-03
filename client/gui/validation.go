package gui

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Valid   bool
	Message string
}

// Valid returns a valid result
func Valid() ValidationResult {
	return ValidationResult{Valid: true}
}

// Invalid returns an invalid result with a message
func Invalid(message string) ValidationResult {
	return ValidationResult{Valid: false, Message: message}
}

// Validator provides validation functions for form fields
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateServerAddress validates server address (hostname:port or IP:port)
func (v *Validator) ValidateServerAddress(address string) ValidationResult {
	if address == "" {
		return Invalid("Server address is required")
	}

	// Split host and port
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return Invalid("Invalid format. Use hostname:port or IP:port")
	}

	// Validate host (hostname or IP)
	if host == "" {
		return Invalid("Hostname or IP address is required")
	}

	// Validate IP address if it looks like one
	if ip := net.ParseIP(host); ip != nil {
		// Valid IP address
	} else {
		// Validate hostname format
		hostnameRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)
		if !hostnameRegex.MatchString(host) {
			return Invalid("Invalid hostname format")
		}
	}

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Invalid("Port must be a number")
	}
	if port < 1 || port > 65535 {
		return Invalid("Port must be between 1 and 65535")
	}

	return Valid()
}

// ValidateFilePath validates that a file path exists
func (v *Validator) ValidateFilePath(path string, checkExists bool) ValidationResult {
	if path == "" {
		return Valid() // Empty is OK for optional fields
	}

	// Clean the path
	path = filepath.Clean(path)

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return Invalid("Path must be absolute")
	}

	// Check if file exists (if required)
	if checkExists {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return Invalid("File or directory does not exist")
		}
	}

	return Valid()
}

// ValidateDirectoryPath validates a directory path
func (v *Validator) ValidateDirectoryPath(path string, checkExists bool) ValidationResult {
	if path == "" {
		return Invalid("Directory path is required")
	}

	result := v.ValidateFilePath(path, checkExists)
	if !result.Valid {
		return result
	}

	// If checking existence, verify it's a directory
	if checkExists {
		info, err := os.Stat(path)
		if err != nil {
			return Invalid("Cannot access directory")
		}
		if !info.IsDir() {
			return Invalid("Path is not a directory")
		}
	}

	return Valid()
}

// ValidatePort validates a port number
func (v *Validator) ValidatePort(portStr string) ValidationResult {
	if portStr == "" {
		return Invalid("Port is required")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Invalid("Port must be a number")
	}

	if port < 1 || port > 65535 {
		return Invalid("Port must be between 1 and 65535")
	}

	return Valid()
}

// ValidatePositiveInt validates a positive integer
func (v *Validator) ValidatePositiveInt(value string, fieldName string) ValidationResult {
	if value == "" {
		return Invalid(fmt.Sprintf("%s is required", fieldName))
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return Invalid(fmt.Sprintf("%s must be a number", fieldName))
	}

	if num <= 0 {
		return Invalid(fmt.Sprintf("%s must be greater than 0", fieldName))
	}

	return Valid()
}

// ValidateIntRange validates an integer within a range
func (v *Validator) ValidateIntRange(value string, fieldName string, min, max int) ValidationResult {
	if value == "" {
		return Invalid(fmt.Sprintf("%s is required", fieldName))
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return Invalid(fmt.Sprintf("%s must be a number", fieldName))
	}

	if num < min || num > max {
		return Invalid(fmt.Sprintf("%s must be between %d and %d", fieldName, min, max))
	}

	return Valid()
}

// ValidateBackupFolders validates a list of backup folders (newline separated)
func (v *Validator) ValidateBackupFolders(foldersText string) ValidationResult {
	if foldersText == "" {
		return Invalid("At least one backup folder is required")
	}

	folders := strings.Split(foldersText, "\n")
	validFolders := 0

	for _, folder := range folders {
		folder = strings.TrimSpace(folder)
		if folder == "" {
			continue // Skip empty lines
		}

		// Check if directory exists
		info, err := os.Stat(folder)
		if err != nil {
			if os.IsNotExist(err) {
				return Invalid(fmt.Sprintf("Folder does not exist: %s", folder))
			}
			return Invalid(fmt.Sprintf("Cannot access folder: %s", folder))
		}

		if !info.IsDir() {
			return Invalid(fmt.Sprintf("Not a directory: %s", folder))
		}

		validFolders++
	}

	if validFolders == 0 {
		return Invalid("At least one valid backup folder is required")
	}

	return Valid()
}

// ValidateExcludePatterns validates exclude patterns (basic validation)
func (v *Validator) ValidateExcludePatterns(patternsText string) ValidationResult {
	// Empty is OK - no exclusions
	if patternsText == "" {
		return Valid()
	}

	patterns := strings.Split(patternsText, "\n")
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		// Validate as glob pattern
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return Invalid(fmt.Sprintf("Invalid pattern: %s", pattern))
		}
	}

	return Valid()
}

// ValidateClientID validates client ID format
func (v *Validator) ValidateClientID(clientID string) ValidationResult {
	if clientID == "" {
		return Invalid("Client ID is required")
	}

	// Client ID should be alphanumeric with dashes/underscores
	validIDRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validIDRegex.MatchString(clientID) {
		return Invalid("Client ID can only contain letters, numbers, dashes, and underscores")
	}

	if len(clientID) < 3 {
		return Invalid("Client ID must be at least 3 characters")
	}

	if len(clientID) > 64 {
		return Invalid("Client ID must be 64 characters or less")
	}

	return Valid()
}

// ValidateChunkSize validates chunk size (in bytes)
func (v *Validator) ValidateChunkSize(sizeStr string) ValidationResult {
	if sizeStr == "" {
		return Invalid("Chunk size is required")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return Invalid("Chunk size must be a number")
	}

	// Minimum 64KB
	if size < 65536 {
		return Invalid("Chunk size must be at least 64KB (65536 bytes)")
	}

	// Maximum 100MB
	if size > 104857600 {
		return Invalid("Chunk size must be at most 100MB (104857600 bytes)")
	}

	return Valid()
}
