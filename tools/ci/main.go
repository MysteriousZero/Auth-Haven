package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var markdownLink = regexp.MustCompile(`\[[^]]*\]\(([^)]+)\)`)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "validate" {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/ci validate")
		os.Exit(2)
	}
	if err := validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func validate() error {
	var failures []string
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		links, err := localLinks(path)
		if err != nil {
			failures = append(failures, err.Error())
			return nil
		}
		for _, link := range links {
			if _, err := os.Stat(link); err != nil {
				failures = append(failures, fmt.Sprintf("%s: missing local link target %s", path, link))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	var yamlFiles []string
	for _, pattern := range []string{".github/ISSUE_TEMPLATE/*.yml", ".github/workflows/*.yml"} {
		matches, _ := filepath.Glob(pattern)
		yamlFiles = append(yamlFiles, matches...)
	}
	for _, path := range yamlFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		var document any
		if err := yaml.Unmarshal(data, &document); err != nil {
			failures = append(failures, fmt.Sprintf("%s: invalid YAML: %v", path, err))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("validation failed:\n- %s", strings.Join(failures, "\n- "))
	}
	return nil
}

func localLinks(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var links []string
	scanner := bufio.NewScanner(file)
	inFence := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		for _, match := range markdownLink.FindAllStringSubmatch(line, -1) {
			target := strings.TrimSpace(strings.SplitN(match[1], "#", 2)[0])
			if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			links = append(links, filepath.Clean(filepath.Join(filepath.Dir(path), target)))
		}
	}
	return links, scanner.Err()
}
