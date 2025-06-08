package pkg

import (
	"os"
	"path/filepath"
	"strings"
)

// ListSourceFiles returns a list of all source code files in the given directory and its subdirectories.
func ListSourceFiles(rootDir, skipList string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a source file by extension
		ext := filepath.Ext(path)
		switch ext {
		case ".go", ".js", ".ts", ".tsx", ".jsx", ".py", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".php", ".rb", ".rs", ".swift", ".kt", ".scala", ".ps1", ".sh":
			if skipList != "" && !strings.Contains(ext, skipList) {
				files = append(files, path)
			} else {
				files = append(files, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
