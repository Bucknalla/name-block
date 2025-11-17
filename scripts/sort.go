package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func sortList(list string) []string {
	words := strings.Split(list, "\n")
	uniqueMap := make(map[string]bool)
	var unique []string

	for _, word := range words {
		trimmed := strings.TrimSpace(strings.ToLower(word))
		if trimmed != "" && !uniqueMap[trimmed] {
			uniqueMap[trimmed] = true
			unique = append(unique, trimmed)
		}
	}

	sort.Strings(unique)
	return unique
}

func main() {
	err := filepath.Walk("data", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".txt") {
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("failed to read file %s: %v\n", path, err)
				return nil
			}

			sorted := sortList(string(data))
			content := strings.Join(sorted, "\n")

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				fmt.Printf("failed to write file %s: %v\n", path, err)
				return nil
			}

			fmt.Printf("Sorted: %s\n", path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
}
