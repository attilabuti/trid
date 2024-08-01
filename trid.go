// This package wraps the TrID command-line tool and parses its output to provide
// structured information about identified file types.
package trid

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrNoFileSpecified is returned when no file path is provided.
	ErrNoFileSpecified = errors.New("no file specified")

	// ErrNumberOfMatches is returned when the specified number of matches is less than 1.
	ErrNumberOfMatches = errors.New("number of matches must be at least 1")

	// ErrNoDefinitions is returned when no TRiD definitions are available.
	ErrNoDefinitions = errors.New("no TRiD definitions available")

	// ErrEmptyDefPackage is returned when a TRiD definition package is empty.
	ErrEmptyDefPackage = errors.New("TRiD definition package is empty")

	// ErrFileNotFound is returned when the specified file cannot be located or accessed.
	ErrFileNotFound = errors.New("file not found")

	// ErrUnknownFileType is returned when TRiD fails to identify the file type.
	ErrUnknownFileType = errors.New("unknown file type")

	// Regular expressions for parsing TRiD output.
	reFileInfo    = regexp.MustCompile(`(?mi)([0-9.]+%)\s+\((\..*?)\)\s+(.*?(?:\s+\([^()]+\))*?)(?:\s+\([^()]+\))?$`)
	reFileDetails = regexp.MustCompile(`(?mi)(Mime type|Related URL|Definition|Remarks)\s*:\s*(.*?)$`)
)

// Trid represents a TrID file identifier instance with specific options.
type Trid struct {
	options Options
}

// Options configures the TrID execution parameters.
type Options struct {
	Cmd         string        // Command to invoke the TrID file identifier.
	Definitions string        // Path to the TrID definitions package.
	Timeout     time.Duration // Maximum duration to wait for TrID execution.
}

// FileType represents detailed information about a file type as identified by TrID.
type FileType struct {
	Extension   string  // File extension (e.g., ".txt", ".pdf").
	Probability float64 // Probability of the file type match, as a percentage (0-100).
	Name        string  // Descriptive name of the file type.
	MimeType    string  // Mime type of the file (e.g., "text/plain", "application/pdf").
	RelatedURL  string  // URL for additional information about the file type.
	Remarks     string  // Additional notes or comments about the file type from TRiD.
	Definition  string  // Name of the TRiD definition XML file for this file type.
}

// NewTrid creates a new Trid instance with the given options.
func NewTrid(opts Options) *Trid {
	if opts.Cmd == "" {
		opts.Cmd = "trid"
	}

	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	return &Trid{opts}
}

// Scan identifies the file type using TRiD, returning a slice of FileType
// structs and an error. It takes a file path and the maximum number of potential
// matches to return.
func (t *Trid) Scan(filePath string, numberOfMatches int) ([]FileType, error) {
	if filePath == "" {
		return nil, ErrNoFileSpecified
	}

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}

		return nil, err
	}

	if numberOfMatches < 1 {
		return nil, ErrNumberOfMatches
	}

	args := []string{"-v", "-n:" + strconv.Itoa(numberOfMatches)}
	if t.options.Definitions != "" {
		args = append(args, "-d:"+t.options.Definitions)
	}
	args = append(args, filePath)

	// Execute TRiD command and capture output
	out, err := execCmd(t.options.Cmd, t.options.Timeout, args...)
	if tridErr := checkTridError(out); tridErr != nil {
		return nil, tridErr
	}

	if err != nil {
		return nil, err
	}

	// Parse the TRiD output
	return parseOutput(out)
}

// parseOutput parses TRiD stdout and returns a slice of FileType structs.
func parseOutput(out string) ([]FileType, error) {
	fileTypes := make([]FileType, 0)

	results := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n\n")
	for _, result := range results {
		fileInfo := reFileInfo.FindStringSubmatch(result)
		if len(fileInfo) != 4 {
			continue
		}

		fileInfo[1] = strings.TrimSpace(strings.Replace(fileInfo[1], "%", "", -1))
		if len(fileInfo[1]) == 0 {
			continue
		}

		probability, err := strconv.ParseFloat(fileInfo[1], 64)
		if err != nil {
			continue
		}

		f := FileType{
			Probability: probability,
			Extension:   strings.ToLower(fileInfo[2]),
			Name:        fileInfo[3],
		}

		fileDetails := reFileDetails.FindAllStringSubmatch(result, -1)
		for _, m := range fileDetails {
			switch m[1] {
			case "Mime type":
				f.MimeType = m[2]
			case "Related URL":
				f.RelatedURL = m[2]
			case "Definition":
				f.Definition = m[2]
			case "Remarks":
				f.Remarks = m[2]
			}
		}

		fileTypes = append(fileTypes, f)
	}

	return fileTypes, nil
}

// checkTridError checks the TrID output for known error messages and returns
// the corresponding error if found.
func checkTridError(out string) error {
	if strings.Contains(out, "you have to specify at least one file to analyze") {
		return ErrNoFileSpecified
	}

	if strings.Contains(out, "No definitions available!") {
		return ErrNoDefinitions
	}

	if strings.Contains(out, "Def package") && strings.Contains(out, "is empty!") {
		return ErrEmptyDefPackage
	}

	if strings.Contains(out, "Error: found no file(s) to analyze!") {
		return ErrFileNotFound
	}

	if strings.Contains(out, "Unknown!") {
		return ErrUnknownFileType
	}

	return nil
}

// execCmd executes a command with a timeout and returns its combined stdout and
// stderr output.
func execCmd(name string, timeout time.Duration, args ...string) (string, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // Ensure resources are cleaned up when the function returns

	// Create the command with the timeout context
	cmd := exec.CommandContext(ctx, name, args...)

	// Execute the command and capture both stdout and stderr
	out, err := cmd.CombinedOutput()

	// Check if the command timed out
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("command timed out: %w", err)
	}

	// Return the output and any execution error
	return string(out), err
}
