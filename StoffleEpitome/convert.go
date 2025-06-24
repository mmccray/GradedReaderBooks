package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
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
		Title:       "Stoffel Epitome",
		Slug:        "stoffel-epitome",
		Author:      "Stoffel",
		Language:    "Greek",
		Description: "A Greek epitome of the life of Jesus, structured in narrative chapters and verses, attributed to Stoffel.",
		Chapters:    []*Chapter{},
	}

	lines := strings.Split(string(content), "\n")
	chapterMap := make(map[int]*Chapter)
	chapterTitles := make(map[int]string)

	verseRegex := regexp.MustCompile(`^([0-9]+)\.([0-9]+)\s+(.*)$`)
	titleRegex := regexp.MustCompile(`^([0-9]+)\.title\s+(.+)$`)
	partMarker := regexp.MustCompile(`^0\.0\s+(.+)$`)
	var pendingPartMarker string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if partMarker.MatchString(line) {
			matches := partMarker.FindStringSubmatch(line)
			pendingPartMarker = matches[1]
			continue
		}
		if titleRegex.MatchString(line) {
			matches := titleRegex.FindStringSubmatch(line)
			chapterNum, _ := strconv.Atoi(matches[1])
			chapterTitles[chapterNum] = matches[2]
			continue
		}
		matches := verseRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			chapterNum, _ := strconv.Atoi(matches[1])
			paraNum, _ := strconv.Atoi(matches[2])
			text := matches[3]
			words := []Word{}
			for _, w := range strings.Fields(text) {
				words = append(words, Word{Word: w, Gloss: ""})
			}
			// If there is a pending part marker, add it as the first paragraph (VerseID 0)
			if chapterMap[chapterNum] == nil {
				chapterMap[chapterNum] = &Chapter{
					Slug:       fmt.Sprintf("chapter-%d", chapterNum),
					Title:      Title{Display: chapterTitles[chapterNum], Gloss: ""},
					TitleImage: "",
					Vocab:      []VocabItem{},
					Questions:  []Question{},
					Content:    []ContentItem{},
				}
				if pendingPartMarker != "" {
					partParagraph := Paragraph{
						VerseID: 0,
						Words:   []Word{{Word: pendingPartMarker, Gloss: ""}},
					}
					chapterMap[chapterNum].Content = append(chapterMap[chapterNum].Content, ContentItem{
						Subtitle:  "",
						Image:     "",
						Paragraph: []Paragraph{partParagraph},
					})
					pendingPartMarker = ""
				}
			}
			paragraph := Paragraph{
				VerseID: paraNum,
				Words:   words,
			}
			contentItem := ContentItem{
				Subtitle:  "",
				Image:     "",
				Paragraph: []Paragraph{paragraph},
			}
			chapterMap[chapterNum].Content = append(chapterMap[chapterNum].Content, contentItem)
		}
	}
	// Collect chapters in order
	chapterNums := []int{}
	for num := range chapterMap {
		chapterNums = append(chapterNums, num)
	}
	sort.Ints(chapterNums)
	for _, num := range chapterNums {
		book.Chapters = append(book.Chapters, chapterMap[num])
	}

	jsonData, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	return string(jsonData), nil
}

func processSection(book *Book, sectionTitle, sectionContent string) {
	var currentChapter *Chapter
	if len(book.Chapters) > 0 {
		currentChapter = book.Chapters[len(book.Chapters)-1]
	}

	switch sectionTitle {
	case "Chapter":
		chapterNumber := len(book.Chapters) + 1
		currentChapter = &Chapter{
			Slug:       fmt.Sprintf("chapter-%d", chapterNumber),
			Title:      Title{},
			TitleImage: "",
			Vocab:      []VocabItem{},
			Questions:  []Question{},
			Content:    []ContentItem{},
		}
		book.Chapters = append(book.Chapters, currentChapter)
		for _, line := range strings.Split(sectionContent, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "Title:") {
				currentChapter.Title.Display = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			} else if strings.HasPrefix(line, "Gloss:") {
				currentChapter.Title.Gloss = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			} else if strings.HasPrefix(line, "TitleImage:") {
				currentChapter.TitleImage = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			}
		}
	case "Vocab":
		if currentChapter == nil {
			fmt.Printf("Warning: [Vocab] section found without a chapter\n")
			return
		}
		hasContent := false
		for _, line := range strings.Split(sectionContent, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) == 3 {
				currentChapter.Vocab = append(currentChapter.Vocab, VocabItem{
					Word:  strings.TrimSpace(parts[0]),
					Gloss: strings.TrimSpace(parts[1]),
					Image: strings.TrimSpace(parts[2]),
				})
				hasContent = true
			}
		}
		if !hasContent {
			chapterNum := len(book.Chapters)
			fmt.Printf("Warning: Empty [Vocab] section in chapter %d\n", chapterNum)
		}
	case "Questions":
		if currentChapter == nil {
			fmt.Printf("Warning: [Questions] section found without a chapter\n")
			return
		}
		hasContent := false
		for _, line := range strings.Split(sectionContent, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) == 2 {
				currentChapter.Questions = append(currentChapter.Questions, Question{
					Question: strings.TrimSpace(parts[0]),
					Answer:   strings.TrimSpace(parts[1]),
				})
				hasContent = true
			}
		}
		if !hasContent {
			chapterNum := len(book.Chapters)
			fmt.Printf("Warning: Empty [Questions] section in chapter %d\n", chapterNum)
		}
	case "Content":
		if currentChapter == nil {
			fmt.Printf("Warning: [Content] section found without a chapter\n")
			return
		}
		lines := strings.Split(sectionContent, "\n")
		verseCounter := 1
		pendingSubtitle := ""
		pendingImage := ""
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "Subtitle:") {
				pendingSubtitle = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			} else if strings.HasPrefix(line, "Image:") {
				pendingImage = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			} else {
				paragraphData := ContentItem{
					Subtitle:  pendingSubtitle,
					Image:     pendingImage,
					Paragraph: []Paragraph{},
				}
				sentenceRegex := regexp.MustCompile(`(?<=[.!?;])\s+`)
				sentences := sentenceRegex.Split(line, -1)
				for _, sentence := range sentences {
					sentence = strings.TrimSpace(sentence)
					if sentence == "" {
						continue
					}
					wordGlossRegex := regexp.MustCompile(`([^()\s]+)(?:\s*\(([^)]+)\))?`)
					matches := wordGlossRegex.FindAllStringSubmatch(sentence, -1)
					words := []Word{}
					for _, match := range matches {
						word := strings.TrimSpace(match[1])
						gloss := ""
						if len(match) > 2 {
							gloss = strings.TrimSpace(match[2])
						}
						words = append(words, Word{Word: word, Gloss: gloss})
					}
					if len(words) > 0 {
						sentenceData := Paragraph{
							VerseID: verseCounter,
							Words:   words,
						}
						paragraphData.Paragraph = append(paragraphData.Paragraph, sentenceData)
						verseCounter++
					}
				}
				if len(paragraphData.Paragraph) == 0 {
					continue
				}
				if paragraphData.Subtitle == "" {
					paragraphData.Subtitle = ""
				}
				if paragraphData.Image == "" {
					paragraphData.Image = ""
				}
				currentChapter.Content = append(currentChapter.Content, paragraphData)
				pendingSubtitle = ""
				pendingImage = ""
			}
		}
	}
}

func main() {
	inputFilePath := "stoffel-epitome.txt"
	outputFilePath := "stoffel-epitome.json"
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
