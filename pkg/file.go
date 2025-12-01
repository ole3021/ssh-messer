package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CleanupResult struct {
	DeletedFiles []string // Deleted file names list
	Errors       []string // Errors list
}

// expandHomeDir Helper function to expand ~ to home directory
func ExpandHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return homeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		return strings.Replace(path, "~", homeDir, 1), nil
	}

	return path, nil
}

func CreateFolderIfNotExist(path string) error {
	expandedPath, err := ExpandHomeDir(path)
	if err != nil {
		return err
	}

	// Check if directory exists
	_, err = os.Stat(expandedPath)
	if err == nil {
		// Directory exists
		return nil
	}

	if !os.IsNotExist(err) {
		// Other errors
		return err
	}

	// Directory does not exist, create it (including all parent directories)
	return os.MkdirAll(expandedPath, 0755)
}

// DeleteFilesByPatternOverDays Delete files in specified directory that match pattern and are older than specified days
// dirPath: directory path (supports ~ expansion)
// pattern: file name pattern (e.g. "ssh-messer#*.log")
// overDays: number of days after which files will be deleted
func DeleteFilesByPatternOverDays(dirPath, pattern string, overDays int) CleanupResult {
	result := CleanupResult{
		DeletedFiles: make([]string, 0),
		Errors:       make([]string, 0),
	}

	expandedDirPath, err := ExpandHomeDir(dirPath)
	if err != nil {
		result.Errors = append(result.Errors, "Failed to expand path: "+err.Error())
		return result
	}

	// Build full matching pattern
	fullPattern := filepath.Join(expandedDirPath, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		result.Errors = append(result.Errors, "Failed to find files: "+err.Error())
		return result
	}

	cutoffTime := time.Now().AddDate(0, 0, -overDays)

	for _, filePath := range matches {
		// Get file information
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			result.Errors = append(result.Errors, "Failed to get file information: "+filepath.Base(filePath)+", error: "+err.Error())
			continue
		}

		// Check if file is older than cutoff time
		if fileInfo.ModTime().Before(cutoffTime) {
			fileName := filepath.Base(filePath)
			if err := os.Remove(filePath); err != nil {
				result.Errors = append(result.Errors, "Failed to delete file: "+fileName+", error: "+err.Error())
			} else {
				result.DeletedFiles = append(result.DeletedFiles, fileName)
			}
		}
	}

	return result
}
