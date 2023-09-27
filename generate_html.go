package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// type Entry struct {
// 	Traditional string            `json:"traditional"`
// 	Simplified  string            `json:"simplified"`
// 	Pinyin      []string          `json:"pinyin"`
// 	Definitions map[string]string `json:"definitions"`
// }

// func createHTML(entry Entry, folderPath string) {
// 	filename := filepath.Join(folderPath, entry.Simplified+".html")
// 	file, err := os.Create(filename)
// 	if err != nil {
// 		fmt.Println("Error creating file:", err)
// 		return
// 	}
// 	defer file.Close()

// 	writer := bufio.NewWriter(file)
// 	fmt.Fprintf(writer, "<h1>%s</h1>\r\n<p>Simplified: %s</p>\r\n<p>Traditional: %s</p>", entry.Simplified, entry.Simplified, entry.Traditional)
// 	for _, pinyin := range entry.Pinyin {
// 		definition, _ := entry.Definitions[pinyin]
// 		fmt.Fprintf(writer, "<p>Pinyin: %s</p>\r\n<p>%s</p>\r\n", pinyin, definition)
// 	}
// 	writer.Flush()
// }

type DictionaryEntry struct {
	Entry      Entry
	Components []Component
}

type Entry struct {
	Traditional string
	Simplified  string
	Pinyin      []string
	Definitions [][]string
}

type Component struct {
	Text       string
	Position   [2]int
	Entry      Entry
	Components []Component
}

func createHTML(dictEntry DictionaryEntry, folderPath string) {
	filename := filepath.Join(folderPath, dictEntry.Entry.Simplified+".html")
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writeEntry(writer, dictEntry.Entry, dictEntry.Components, true) // the root entry is not a component
	writer.Flush()
}

func writeEntry(writer *bufio.Writer, entry Entry, components []Component, isRoot bool) {
	if isRoot {
		fmt.Fprintf(writer, "<h1>%s</h1>\r\n<p>Simplified: %s</p>\r\n<p>Traditional: %s</p>", entry.Simplified, entry.Simplified, entry.Traditional)
	} else {
		// Use <details> and <summary> for components
		fmt.Fprintf(writer, "<details><summary>%s</summary>\r\n<p>Simplified: %s</p>\r\n<p>Traditional: %s</p>", entry.Simplified, entry.Simplified, entry.Traditional)
	}

	for i, pinyin := range entry.Pinyin {
		definition := entry.Definitions[i]
		definitionStr := strings.Join(definition, "; ")
		fmt.Fprintf(writer, "<p>Pinyin: %s</p>\r\n<p>Definition: %s</p>\r\n", pinyin, definitionStr)
	}

	for _, component := range components {
		writeEntry(writer, component.Entry, component.Components, false) // recursive call to handle nested components
	}

	if !isRoot {
		fmt.Fprintf(writer, "</details>")
	}
}

func deleteFilesInFolder(folderPath string) {
	files, err := filepath.Glob(filepath.Join(folderPath, "*.html"))
	if err != nil {
		fmt.Println("Error fetching files:", err)
		return
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			fmt.Println("Error deleting file:", err)
		}
	}
}

func printProgress(progress, total int) {
	// Calculate percentage
	percent := float64(progress) / float64(total) * 100
	// Calculate number of blocks to represent progress
	blocks := int(percent / 2)

	// Print the progress bar
	fmt.Printf("\r[")
	for i := 0; i < blocks; i++ {
		fmt.Print("=")
	}
	for i := 0; i < 50-blocks; i++ {
		fmt.Print(" ")
	}
	fmt.Printf("] %3.0f%%", percent)

	if progress == total {
		fmt.Println() // Print a newline at 100%
	}
}

func main() {
	buildFolder := "docs"

	err := os.MkdirAll(buildFolder, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating build folder:", err)
		return
	}

	// Delete existing HTML files
	deleteFilesInFolder(buildFolder)

	indexFile, err := os.Create(filepath.Join(buildFolder, "index.html"))
	if err != nil {
		fmt.Println("Error creating index file:", err)
		return
	}
	defer indexFile.Close()

	indexWriter := bufio.NewWriter(indexFile)
	fmt.Fprint(indexWriter, `<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<style>
		.grid-container {
			display: flex;
			flex-wrap: wrap;
		}
		.grid-item {
			width: 100px;
			height: 100px;
			display: flex;
			align-items: center;
			justify-content: center;
			border: 1px solid black;
			flex-grow: 1;
			text-align: center;
		}
		@media (max-width: 600px) {
			.grid-item {
				width: 50px;
				height: 50px;
			}
		}
	</style>
</head>
<body>
	<div class="grid-container">
`)

	file, err := os.Open("cedict_with_components.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var entries map[string]DictionaryEntry

	err = json.Unmarshal(bytes, &entries)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	total := len(entries) // Get the total number of entries
	var progress int
	var printMutex sync.Mutex

	// Create a semaphore channel to limit concurrent goroutines
	sem := make(chan bool, 64) // 20 is the number of concurrent goroutines allowed

	var wg sync.WaitGroup

	for _, dictEntry := range entries {
		wg.Add(1)
		sem <- true // Acquire a token
		go func(e DictionaryEntry) {
			defer wg.Done()
			createHTML(e, buildFolder)

			progress++
			printMutex.Lock()
			printProgress(progress, total)
			printMutex.Unlock()

			<-sem // Release the token
		}(dictEntry)

		// Add the link to the index file
		fmt.Fprintf(indexWriter, `<a class="grid-item" href="%s">%s</a>`+"\n", url.QueryEscape(dictEntry.Entry.Simplified), dictEntry.Entry.Simplified)
	}

	fmt.Fprint(indexWriter, `
	</div>
</body>
</html>
`)
	indexWriter.Flush()

	wg.Wait()
}
