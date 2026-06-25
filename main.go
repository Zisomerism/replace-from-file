package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobwas/glob"
)

var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func getenvStr(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, ErrEnvVarEmpty
	}
	return v, nil
}

func check(e error) {
	if e != nil {
		log.Fatalf("error: %v", e)
	}
}

func loadMappings(sourceDir string) (map[string]string, error) {
	info, err := os.Stat(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("source_dir %q: %w", sourceDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source_dir %q is not a directory", sourceDir)
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("reading source_dir %q: %w", sourceDir, err)
	}

	mappings := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if stem == "" {
			continue
		}

		if _, exists := mappings[stem]; exists {
			return nil, fmt.Errorf("duplicate stem %q from source_dir", stem)
		}

		content, err := os.ReadFile(filepath.Join(sourceDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %q: %w", entry.Name(), err)
		}

		mappings[stem] = string(content)
	}

	return mappings, nil
}

func sortedStems(mappings map[string]string) []string {
	stems := make([]string, 0, len(mappings))
	for stem := range mappings {
		stems = append(stems, stem)
	}

	sort.Slice(stems, func(i, j int) bool {
		return len(stems[i]) > len(stems[j])
	})

	return stems
}

func listFiles(include string, exclude string, sourceDir string) ([]string, error) {
	sourceDirPattern := filepath.ToSlash(sourceDir) + "/**"
	fileList := []string{}

	err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if doesFileMatch(path, include, exclude) && !doesFileMatch(path, sourceDirPattern, "") {
			fileList = append(fileList, path)
		}
		return nil
	})

	return fileList, err
}

func doesFileMatch(path string, include string, exclude string) bool {
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		includeGlob := glob.MustCompile(include)
		matchPath := filepath.ToSlash(path)
		if !includeGlob.Match(matchPath) {
			return false
		}
		if exclude == "" {
			return true
		}
		excludeGlob := glob.MustCompile(exclude)
		return !excludeGlob.Match(matchPath)
	}
	return false
}

func applyReplacements(path string, mappings map[string]string, stems []string) (bool, error) {
	read, readErr := os.ReadFile(path)
	if readErr != nil {
		return false, readErr
	}

	newContents := string(read)
	for _, stem := range stems {
		replacement := mappings[stem]
		if strings.Contains(newContents, stem) {
			newContents = strings.ReplaceAll(newContents, stem, replacement)
		}
	}

	if newContents == string(read) {
		return false, nil
	}

	writeErr := os.WriteFile(path, []byte(newContents), 0)
	if writeErr != nil {
		return false, writeErr
	}

	return true, nil
}

func setGithubEnvOutput(key string, value int) {
	outputFilename := os.Getenv("GITHUB_OUTPUT")
	f, err := os.OpenFile(outputFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "%s=%d\n", key, value); err != nil {
		log.Println(err)
	}
}

func main() {
	sourceDir, sourceDirErr := getenvStr("INPUT_SOURCE_DIR")
	if sourceDirErr != nil {
		log.Fatal("gha-txt-fill: expected with.source_dir to be a string")
	}

	include, includeErr := getenvStr("INPUT_INCLUDE")
	if includeErr != nil {
		include = "**"
	}

	exclude, excludeErr := getenvStr("INPUT_EXCLUDE")
	if excludeErr != nil {
		exclude = ".git/**"
	}

	mappings, mappingsErr := loadMappings(sourceDir)
	check(mappingsErr)

	stems := sortedStems(mappings)

	files, filesErr := listFiles(include, exclude, sourceDir)
	check(filesErr)

	modifiedCount := 0

	for _, path := range files {
		modified, replaceErr := applyReplacements(path, mappings, stems)
		check(replaceErr)

		if modified {
			modifiedCount++
		}
	}

	setGithubEnvOutput("modifiedFiles", modifiedCount)
}
