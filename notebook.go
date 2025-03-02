package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titler = cases.Title(language.AmericanEnglish)

//go:embed page.tmpl.html
var templateFileSource string
var funcMap = template.FuncMap{
	"ToTitle": func(s string) string { return titler.String(s) },
}
var templateFile = template.Must(template.New("outputPage").Funcs(funcMap).Parse(templateFileSource))

func makeNotebookFile(allNotes []Note, frontmatterWanted []string, includeUnfinished bool, frontPageFile string, outputFile string) {

	wanted_entries := filterFilesForThisNotebook(allNotes, includeUnfinished)
	slices.SortFunc(wanted_entries, sortEntries)

	entries := []RenderedEntry{}
	frontmatterNotes := []RenderedEntry{}
	errs := []error{}
	byFocus := map[string][]Note{}

	for _, metadata := range wanted_entries {
		buf := bytes.NewBuffer([]byte{})

		err := NotebookRender().Render(buf, metadata.Src, metadata.Doc)
		if err != nil {
			errs = append(errs, err)
		}

		gradientHR := metadata.GradientHR()

		if slices.Contains(metadata.ProcessSteps, "frontmatter") {
			frontmatterNotes = append(frontmatterNotes, RenderedEntry{
				Data:       metadata,
				Html:       buf.String(),
				GradientHR: gradientHR,
			})
			continue
		}

		// Sort by focus
		if l, exists := byFocus[metadata.Focus]; exists {
			byFocus[metadata.Focus] = append(l, metadata)
		} else {
			byFocus[metadata.Focus] = []Note{metadata}
		}

		entries = append(entries, RenderedEntry{
			Data:       metadata,
			Html:       buf.String(),
			GradientHR: gradientHR,
		})
	}

	if len(errs) > 0 {
		fmt.Println(errs)
	}

	focusList := []FocusGroup{}
	for focus, entries := range byFocus {
		slices.SortFunc(entries, func(a, b Note) int {
			return a.Date.Compare(b.Date)
		})
		if focus != "Frontmatter" {
			focusList = append(focusList, FocusGroup{
				Focus:    focus,
				Entries:  entries,
				Notebook: entries[0].Notebook,
			})
		}
	}

	slices.SortFunc(focusList, func(a, b FocusGroup) int {
		return a.Entries[0].Date.Compare(b.Entries[0].Date)
	})

	for j, entry := range entries {
		focus := entry.Data.Focus
		neighbors := byFocus[focus]
		i := slices.IndexFunc(neighbors, func(n Note) bool {
			return n.Location == entry.Data.Location
		})

		if i > 0 {
			entries[j].Data.PrevInFocus = &neighbors[i-1]
		}
		if i < len(neighbors)-1 {
			entries[j].Data.NextInFocus = &neighbors[i+1]
		}
	}

	orderedFrontmatterNotes := []RenderedEntry{}
	for _, name := range frontmatterWanted {
		found := false
		for _, note := range frontmatterNotes {
			if note.Data.Title == name {
				orderedFrontmatterNotes = append(orderedFrontmatterNotes, note)
				found = true
			}
		}
		if !found {
			log.Printf("Couldnt find entry '%s' for notebook", name)
		}
	}

	writeNotebookHTMLToFile(outputFile, frontPageFile, orderedFrontmatterNotes, entries, focusList)
}

func sortEntries(a, b Note) int {
	cmp := a.Date.Compare(b.Date)
	if cmp != 0 {
		return cmp
	}

	stepPriority := map[string]int{
		"context":          0,
		"identify-problem": 1,
		"game-analysis":    2,
		"brainstorm":       3,
		"select-best":      4,
		"update":           5,
		"test-result":      6,
		"other":            7,
	}

	defaultPriority := 10

	aPriority := defaultPriority
	if len(a.ProcessSteps) > 0 {
		if p, ok := stepPriority[a.ProcessSteps[0]]; ok {
			aPriority = p
		}
	}

	bPriority := defaultPriority
	if len(b.ProcessSteps) > 0 {
		if p, ok := stepPriority[b.ProcessSteps[0]]; ok {
			bPriority = p
		}
	}

	if aPriority != bPriority {
		return aPriority - bPriority
	}

	return strings.Compare(a.Title, b.Title)
}

func writeNotebookHTMLToFile(filepath string, frontPagePath string, frontmatter []RenderedEntry, entries []RenderedEntry, focusList []FocusGroup) {
	f, err := os.Create(path.Join(tmpDir, filepath))
	must(err)
	err = f.Truncate(0)
	must(err)
	defer f.Close()

	frontpage := ""
	if frontPagePath != "" {
		bs, err := os.ReadFile(frontPagePath)
		if err != nil {
			log.Fatalf("Failed to read front page path: %v", err)
		}
		frontpage = string(bs)
	}

	err = templateFile.Execute(f, Notebook{
		Date:        time.Now(),
		Frontmatter: frontmatter,
		FrontPage:   string(frontpage),
		Entries:     entries,
		ByFocus:     focusList,
	})
	must(err)

}

func getAllFilesInDirectory(root string) []string {
	files := []string{}
	f := func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, path.Join(root, filename))
		return nil
	}

	err := fs.WalkDir(os.DirFS(root), ".", f)
	must(err)
	return files
}
