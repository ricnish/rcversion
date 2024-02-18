package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func replaceVersion(line string, version string) (string, bool) {
	versionComma := strings.ReplaceAll(version, ".", ",")
	if strings.Index(line, " FILEVERSION ") == 0 {
		return " FILEVERSION " + versionComma, true
	}

	if strings.Index(line, " PRODUCTVERSION ") == 0 {
		return " PRODUCTVERSION " + versionComma, true
	}

	if strings.Contains(line, "VALUE \"FileVersion\"") {
		return "            VALUE \"FileVersion\", \"" + version + "\"", true
	}

	if strings.Contains(line, "VALUE \"ProductVersion\"") {
		return "            VALUE \"ProductVersion\", \"" + version + "\"", true
	}

	return line, false
}

func readContent(rcpath string, isUtf16 bool) []string {
	var lines []string

	file, err := os.Open(rcpath)
	if err != nil {
		fmt.Printf("Could not open the file %s\n", rcpath)
		return lines
	}

	defer file.Close()

	var scanner *bufio.Scanner

	if isUtf16 {
		// Make an tranformer that converts MS-Win default to UTF8:
		win16be := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
		// Make a transformer that is like win16be, but abides by BOM:
		utf16bom := unicode.BOMOverride(win16be.NewDecoder())
		// Make a Reader that uses utf16bom:
		unicodeReader := transform.NewReader(file, utf16bom)
		scanner = bufio.NewScanner(unicodeReader)
	} else {
		// Default to UTF8/ANSI
		scanner = bufio.NewScanner(file)
	}

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	return lines
}

func detectEncoding(rcpath string) *chardet.Result {
	file, err := os.Open(rcpath)
	if err != nil {
		fmt.Printf("Could not open the file: %s\n", rcpath)
		os.Exit(1)
	}

	textDetector := chardet.NewTextDetector()
	buffer := make([]byte, 32<<10)
	size, _ := io.ReadFull(file, buffer)
	input := buffer[:size]
	textEncoding, err := textDetector.DetectBest(input)
	file.Close()

	if err != nil {
		fmt.Printf("Could not detect encoding from file %s\n", rcpath)
		os.Exit(1)
	}

	return textEncoding
}

func readVersion(rcpath string) string {
	if path.Ext(rcpath) != ".rc" {
		fmt.Printf("Is not a .rc file %s\n", rcpath)
		os.Exit(1)
	}

	var textEncoding *chardet.Result = detectEncoding(rcpath)

	var isUtf16 bool = textEncoding.Charset == "UTF-16LE"

	lines := readContent(rcpath, isUtf16)
	for _, line := range lines {
		if strings.Contains(line, "VALUE \"FileVersion\"") {
			strs := strings.Split(line, ", ")
			str := strs[1]
			return strings.Trim(str, "\"")
		}
	}

	return ""
}

func writeContent(tmpfile string, lines []string, isUtf16 bool) bool {
	newfile, err := os.Create(tmpfile)

	if err != nil {
		fmt.Printf("Could not create the file %s\n", tmpfile)
		return false
	}

	defer newfile.Close()

	if isUtf16 {
		var bytes [2]byte
		const BOM = '\ufffe' //LE. for BE '\ufeff'
		data := strings.Join(lines, "\r\n") + "\r\n"
		bytes[0] = BOM >> 8
		bytes[1] = BOM & 255

		newfile.Write(bytes[0:])
		runes := utf16.Encode([]rune(data))
		for _, r := range runes {
			bytes[1] = byte(r >> 8)
			bytes[0] = byte(r & 255)
			newfile.Write(bytes[0:])
		}
	} else {
		w := bufio.NewWriter(newfile)
		for _, line := range lines {
			// Add line with CR LF ending
			fmt.Fprintf(w, "%s\r\n", line)
		}
		w.Flush()
	}

	return true
}

func changeFileVersion(rcpath string, version string) bool {
	if path.Ext(rcpath) != ".rc" {
		fmt.Printf("Is not a .rc file %s\n", rcpath)
		return false
	}

	var textEncoding *chardet.Result = detectEncoding(rcpath)
	var isUtf16 bool = textEncoding.Charset == "UTF-16LE"
	var hasChanged bool = false

	lines := readContent(rcpath, isUtf16)

	for i, line := range lines {
		tmpline, haveChanges := replaceVersion(line, version)
		if haveChanges {
			lines[i] = tmpline
			hasChanged = true
		}
	}

	if !hasChanged {
		return false
	}

	tmpfile := strings.Replace(rcpath, ".rc", ".rcbk", 1)

	if !writeContent(tmpfile, lines, isUtf16) {
		fmt.Printf("Could not create the file %s\n", tmpfile)
		return false
	}

	if err := os.Remove(rcpath); err != nil {
		fmt.Printf("Could not delete the original file: %s\n", rcpath)
		return false
	}

	if err := os.Rename(tmpfile, rcpath); err != nil {
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
		var rcfiles []string

		err = filepath.Walk(rcpath, func(walkpath string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error reading path %s\n", walkpath)
				return nil
			}

			if !info.IsDir() && filepath.Ext(walkpath) == ".rc" {
				rcfiles = append(rcfiles, walkpath)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error walking the path %s", rcpath)
		}

		chRes := make(chan bool)
		for _, rcfile := range rcfiles {
			go func(f string, v string) {
				res := changeFileVersion(f, v)
				if res {
					fmt.Println(f)
				}
				chRes <- res
			}(rcfile, version)
		}

		for i := 0; i < len(rcfiles); i++ {
			if <-chRes {
				count++
			}
		}
	} else {
		// Just change a single file
		if changeFileVersion(rcpath, version) {
			count++
		}
	}

	return count
}
