package pkg

import (
	"os"
	"path/filepath"
	"strings"
)

// BuildSourceFileList returns a list of all source code files in the given directory and its subdirectories.
func BuildSourceFileList(rootDir, skipList, overrideList string) ([]string, error) {
	var files []string
	skipList = strings.ToLower(skipList)
	if skipList != "" {
		println("Skip list provided:", skipList)
	}
	overrideList = strings.ToLower(overrideList)
	if overrideList != "" {
		println("Override list provided:", overrideList)
	}

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

		if overrideList == "" {
			switch ext {
			case ".go", ".js", ".ts", ".tsx", ".jsx", ".py", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".php", ".rb", ".rs", ".swift", ".kt", ".scala", ".ps1", ".sh":
				if skipList != "" && !strings.Contains(skipList, ext) {
					files = append(files, path)
					println("Adding file to list:", path)
				} else if skipList == "" {
					files = append(files, path)
				}
			}
		} else {
			if ext != "" && strings.Contains(overrideList, ext) {
				println("Adding file to list:", path)
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
