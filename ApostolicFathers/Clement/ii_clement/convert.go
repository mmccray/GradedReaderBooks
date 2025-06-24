package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Book struct {
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Author      string     `json:"author"`
	Language    string     `json:"language"`
	Description string     `json:"description"`
	CoverImage  string     `json:"coverImage,omitempty"`
	Chapters    []*Chapter `json:"chapters"`
}

type Chapter struct {
	Slug       string        `json:"slug"`
	Title      Title         `json:"title"`
	TitleImage string        `json:"titleImage"`
	Vocab      []VocabItem   `json:"vocab"`
	Questions  []Question    `json:"questions"`
	Content    []ContentItem `json:"content"`
}

type Title struct {
	Display string `json:"display"`
	Gloss   string `json:"gloss"`
}

type VocabItem struct {
	Word  string `json:"word"`
	Gloss string `json:"gloss"`
	Image string `json:"image"`
}

type Question struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type ContentItem struct {
	Subtitle  string      `json:"subtitle,omitempty"`
	Image     string      `json:"image,omitempty"`
	Paragraph []Paragraph `json:"paragraph"`
}

type Paragraph struct {
	VerseID int    `json:"verse_id"`
	Words   []Word `json:"words"`
}

type Word struct {
	Word  string `json:"word"`
	Gloss string `json:"gloss"`
}

func parseTextToJSON(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	book := &Book{
		Title:    "Second Epistle of Clement to the Corinthians",
		Slug:     "ii_clement",
		Author:   "Clement of Rome",
		Language: "Greek",
		Chapters: []*Chapter{},
	}

	lines := strings.Split(string(content), "\n")
	chapter := &Chapter{
		Slug:       "chapter-1",
		Title:      Title{Display: "Second Epistle of Clement to the Corinthians", Gloss: ""},
		TitleImage: "",
		Vocab:      []VocabItem{},
		Questions:  []Question{},
		Content:    []ContentItem{},
	}

	verseRegex := regexp.MustCompile(`^([0-9]+\.[0-9]+)\s+(.*)$`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := verseRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			verseID := matches[1]
			text := matches[2]
			words := []Word{}
			for _, w := range strings.Fields(text) {
				words = append(words, Word{Word: w, Gloss: ""})
			}
			paragraph := Paragraph{
				VerseID: 0,
				Words:   words,
			}
			if parts := strings.Split(verseID, "."); len(parts) == 2 {
				if v, err := strconv.Atoi(parts[0] + parts[1]); err == nil {
					paragraph.VerseID = v
				}
			}
			contentItem := ContentItem{
				Subtitle:  "",
				Image:     "",
				Paragraph: []Paragraph{paragraph},
			}
			chapter.Content = append(chapter.Content, contentItem)
		}
	}
	book.Chapters = append(book.Chapters, chapter)

	jsonData, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	return string(jsonData), nil
}

func main() {
	inputFilePath := "ii_clement.txt"
	outputFilePath := "ii_clement.json"
	result, err := parseTextToJSON(inputFilePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	err = os.WriteFile(outputFilePath, []byte(result), 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Processing complete. Output saved to %s\n", outputFilePath)
}
