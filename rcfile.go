package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func replaceVersion(line string, version string) string {
	versionComma := strings.ReplaceAll(version, ".", ",")
	if strings.Index(line, " FILEVERSION ") == 0 {
		return " FILEVERSION " + versionComma
	}

	if strings.Index(line, " PRODUCTVERSION ") == 0 {
		return " PRODUCTVERSION " + versionComma
	}

	if strings.Contains(line, "VALUE \"FileVersion\"") {
		return "            VALUE \"FileVersion\", \"" + version + "\""
	}

	if strings.Contains(line, "VALUE \"ProductVersion\"") {
		return "            VALUE \"ProductVersion\", \"" + version + "\""
	}

	return line
}

func readVersion(rcpath string) string {
	if path.Ext(rcpath) != ".rc" {
		fmt.Printf("Is not a .rc file %s\n", rcpath)
		os.Exit(1)
	}

	file, err := os.Open(rcpath)
	if err != nil {
		fmt.Printf("Could not open the file: %s\n", rcpath)
		os.Exit(1)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "VALUE \"FileVersion\"") {
			strs := strings.Split(line, ", ")
			str := strs[1]
			return strings.Trim(str, "\"")
		}
	}

	return ""
}

func changeFileVersion(rcpath string, version string) bool {
	if path.Ext(rcpath) != ".rc" {
		fmt.Printf("Is not a .rc file %s\n", rcpath)
		return false
	}

	file, err := os.Open(rcpath)
	if err != nil {
		fmt.Printf("Could not open the file %s\n", rcpath)
		return false
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	file.Close()

	tmpfile := strings.Replace(rcpath, ".rc", ".rcbk", 1)

	newfile, err := os.Create(tmpfile)

	if err != nil {
		fmt.Printf("Could not create the file %s\n", tmpfile)
		return false
	}

	w := bufio.NewWriter(newfile)
	for _, line := range lines {
		// Process each line for version change
		newline := replaceVersion(line, version)
		// Add line with CR LF ending
		fmt.Fprintf(w, "%s\r\n", newline)
	}
	w.Flush()
	newfile.Close()

	err = os.Remove(rcpath)
	if err != nil {
		fmt.Printf("Could not delete the original file: %s\n", rcpath)
		return false
	}

	err = os.Rename(tmpfile, rcpath)
	if err != nil {
		fmt.Printf("Could not rename the new file: %s\n", tmpfile)
		return false
	}
	return true
}

func changeVersion(rcpath string, version string) int {
	var count int = 0

	fileinfo, err := os.Stat(rcpath)
	if err != nil {
		fmt.Printf("Could not find the path %s\n", rcpath)
		return count
	}

	if fileinfo.IsDir() {
		// If path is a folder, find resource files recursively
		err = filepath.Walk(rcpath, func(walkpath string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error reading path %s", walkpath)
				return nil
			}

			if path.Ext(info.Name()) == ".rc" {
				if changeFileVersion(walkpath, version) {
					fmt.Println(walkpath)
					count++
				}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Error walking the path %s", rcpath)
		}

		return count
	} else {
		// Just change a single file
		if changeFileVersion(rcpath, version) {
			count++
		}
	}

	return count
}
