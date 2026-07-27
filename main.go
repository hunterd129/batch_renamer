package main

import (
	"bufio"
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

type RenamePair struct {
	oldName string
	newName string
	srcPath string
	dstPath string
}

func askConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nProceed with changes? [Y/n]: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(input)
	if input == "" || strings.HasPrefix(strings.ToLower(input), "y") {
		return true
	}

	return false
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "batren - Batch Image Renaming Utility\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, " batren -t <path> -p <prefix> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, " -t, --target <path>		Path to target directory containing images\n")
		fmt.Fprintf(os.Stderr, " -p, --prefix <string>	Prefix to apply to renamed files\n")
		fmt.Fprintf(os.Stderr, " -y, --no-confirm		Skip confirmation prompt and execute immediately\n")
		fmt.Fprintf(os.Stderr, " -h, --help				Display options & usage\n")
	}

	var targetFolder, targetFolderShort string
	var prefix, prefixShort string

	flag.StringVar(&targetFolder, "target", "", "Path to the target directory")
	flag.StringVar(&targetFolderShort, "t", "", "Path to the target directory (shorthand)")
	flag.StringVar(&prefix, "prefix", "", "Prefix for renamed files")
	flag.StringVar(&prefixShort, "p", "", "Prefix for renamed files (shorthand)")
	noConfirm := flag.Bool("no-confirm", false, "Skip confirmation prompt and proceed automatically")
	flag.BoolVar(noConfirm, "y", false, "Skip confirmation prompt(short alias)")

	for _, arg := range os.Args[1:] {
		if arg == "--" {
			break
		}

		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			trimmed := strings.TrimPrefix(arg, "-")
			optName, _, _ := strings.Cut(trimmed, "=")

			if len(optName) > 1 {
				fmt.Printf("Error: long option '%s' must use '--' (e.g. --%s)\n", arg, optName)
				os.Exit(1)
			}
		}
	}

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

	var pairs []RenamePair

	fmt.Printf("Normalizing %d files in: %s\n", totalFiles, resolvedPath)
	fmt.Printf("Applying prefix: %s\n", prefix)
	fmt.Println("Changes to be made:")

	for count, file := range files {
		index := count + 1
		finalName := fmt.Sprintf("%s-%0*d%s", prefix, padding, index, file.ext)

		pair := RenamePair{
			oldName: file.name,
			newName: finalName,
			srcPath: filepath.Join(resolvedPath, file.name),
			dstPath: filepath.Join(resolvedPath, finalName),
		}

		pairs = append(pairs, pair)
		fmt.Printf(" %s -> %s\n", pair.oldName, pair.newName)
	}

	if !*noConfirm {
		if !askConfirmation() {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	fmt.Println("\nRenaming files...")
	for _, pair := range pairs {
		if pair.oldName != pair.newName {
			err := os.Rename(pair.srcPath, pair.dstPath)
			if err != nil {
				fmt.Printf("Failed to rename %s: %v\n", pair.oldName, err)
				continue
			}
			fmt.Printf("Renamed: %s -> %s\n", pair.oldName, pair.newName)
		}
	}

	fmt.Println("Changes complete.")
}
