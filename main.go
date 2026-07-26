package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var validExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func main() {
	var targetFolder, targetFolderShort string
	var prefix, prefixShort string

	flag.StringVar(&targetFolder, "target", "", "Path to the target directory")
	flag.StringVar(&targetFolderShort, "t", "", "Path to the target directory (shorthand)")
	flag.StringVar(&prefix, "prefix", "", "Prefix for renamed files")
	flag.StringVar(&prefixShort, "p", "", "Prefix for renamed files (shorthand)")
	flag.Parse()

	if targetFolder == "" {
		targetFolder = targetFolderShort
	}
	if prefix == "" {
		prefix = prefixShort
	}

	if targetFolder == "" || prefix == "" {
		fmt.Println("Error: Both -target (-t) and -prefix (-p) are required.")
		flag.Usage()
		os.Exit(1)
	}

	resolvedPath, err := filepath.Abs(targetFolder)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(resolvedPath)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: the folder '%s' does not exist or is not a directory.\n", targetFolder)
		os.Exit(1)
	}

	entries, err := os.ReadDir(resolvedPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}

	type fileItem struct {
		name    string
		modTime int64
		ext     string
	}

	var files []fileItem
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if validExts[ext] {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, fileItem{
				name:    entry.Name(),
				modTime: info.ModTime().UnixNano(),
				ext:     ext,
			})
		}
	}

	totalFiles := len(files)
	if totalFiles == 0 {
		fmt.Printf("No compatible images found in %s.\n", resolvedPath)
		os.Exit(0)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime < files[j].modTime
	})

	padding := 2
	if totalFiles > 99 {
		padding = 3
	}

	fmt.Printf("Normalizing %d files in: %s\n", totalFiles, resolvedPath)
	fmt.Printf("Applying prefix: %s\n", prefix)

	for count, file := range files {
		index := count + 1
		finalName := fmt.Sprintf("%s-%0*d%s", prefix, padding, index, file.ext)

		sourcePath := filepath.Join(resolvedPath, file.name)
		destPath := filepath.Join(resolvedPath, finalName)

		if file.name != finalName {
			err := os.Rename(sourcePath, destPath)
			if err != nil {
				fmt.Printf("Failed to rename %s: %v\n", file.name, err)
				continue
			}
			fmt.Printf("Renamed: %s -> %s\n", file.name, finalName)
		}
	}

	fmt.Println("Ingestion pipeline complete.")
}
