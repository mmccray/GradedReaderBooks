package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Word struct {
	Word  string `json:"word"`
	Gloss string `json:"gloss"`
}

type Paragraph struct {
	VerseID int    `json:"verse_id"`
	Words   []Word `json:"words"`
}

type Content struct {
	Subtitle  string      `json:"subtitle"`
	Paragraph []Paragraph `json:"paragraph"`
}

type Title struct {
	Display string `json:"display"`
	Gloss   string `json:"gloss"`
}

type Chapter struct {
	Slug       string    `json:"slug"`
	Title      Title     `json:"title"`
	TitleImage string    `json:"titleImage"`
	Vocab      []string  `json:"vocab"`
	Questions  []string  `json:"questions"`
	Content    []Content `json:"content"`
}

type Book struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Author      string    `json:"author"`
	Language    string    `json:"language"`
	Description string    `json:"description"`
	CoverImage  string    `json:"coverImage"`
	Restricted  bool      `json:"restricted"`
	Chapters    []Chapter `json:"chapters"`
}

func main() {
	inputFile := "1-apology.txt"
	outputFile := "1-apology.json"

	book := Book{
		ID:          "1-apology",
		Title:       "1 Apology",
		Slug:        "1-apology",
		Author:      "Justin Martyr",
		Language:    "greek",
		Description: "The First Apology of Justin Martyr.",
		CoverImage:  "1-apology.png",
		Restricted:  false,
		Chapters:    []Chapter{},
	}

	file, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	chapterMap := make(map[string]*Chapter)
	chapterOrder := []string{}
	verseRe := regexp.MustCompile(`^(\d+)\.(\d+)\s+(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		matches := verseRe.FindStringSubmatch(line)
		if len(matches) == 4 {
			chapterNum := matches[1]
			verseNum := matches[2]
			text := matches[3]
			chapterSlug := fmt.Sprintf("chapter-%s", chapterNum)
			verseID := 0
			fmt.Sscanf(verseNum, "%d", &verseID)

			// Create chapter if not exists
			if _, ok := chapterMap[chapterSlug]; !ok {
				chapter := &Chapter{
					Slug:       chapterSlug,
					Title:      Title{Display: fmt.Sprintf("Chapter %s", chapterNum), Gloss: ""},
					TitleImage: "",
					Vocab:      []string{},
					Questions:  []string{},
					Content:    []Content{},
				}
				chapterMap[chapterSlug] = chapter
				chapterOrder = append(chapterOrder, chapterSlug)
			}

			words := []Word{}
			for _, w := range strings.Fields(text) {
				words = append(words, Word{Word: w, Gloss: ""})
			}
			paragraph := Paragraph{VerseID: verseID, Words: words}
			content := Content{Subtitle: "", Paragraph: []Paragraph{paragraph}}
			chapter := chapterMap[chapterSlug]
			chapter.Content = append(chapter.Content, content)
		}
	}

	for _, slug := range chapterOrder {
		book.Chapters = append(book.Chapters, *chapterMap[slug])
	}

	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(book); err != nil {
		panic(err)
	}

	fmt.Println("Conversion complete! Output written to", outputFile)
}
