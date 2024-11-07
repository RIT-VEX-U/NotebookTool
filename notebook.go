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

//go:embed page.tmpl.html
var templateFileSource string

var templateFile = template.Must(template.New("outputPage").Parse(templateFileSource))

func makeNotebookFile(notebook string, allNotes []Note, frontmatterWanted []string) {

	wanted_entries := filterFilesForThisNotebook(allNotes, notebook)

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

		// Rout to frontmatter or regular entry
		if slices.Contains(metadata.ProcessSteps, "frontmatter") {
			frontmatterNotes = append(frontmatterNotes, RenderedEntry{
				Data: metadata,
				Html: buf.String(),
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
			Data: metadata,
			Html: buf.String(),
		})

	}
	if len(errs) > 0 {
		fmt.Println(errs)
	}

	// Get focii
	focusList := []FocusGroup{}
	for focus, entries := range byFocus {
		// Sort group by date
		slices.SortFunc(entries, func(a, b Note) int {
			return a.Date.Compare(b.Date)
		})
		if focus != "Frontmatter" {
			focusList = append(focusList, FocusGroup{
				Focus:   focus,
				Entries: entries,
			})
		}
	}

	// Sort focus index by focus name
	slices.SortFunc(focusList, func(a, b FocusGroup) int {
		return a.Entries[0].Date.Compare(b.Entries[0].Date)
	})

	// Find neighbours
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
			log.Printf("Couldnt find entry '%s' for %s notebook", name, notebook)
		}
	}

	writeNotebookHTMLToFile(notebook, orderedFrontmatterNotes, entries, focusList)
}

func sortEntries(a, b Note) int {
	cmp := a.Date.Compare(b.Date)
	if cmp != 0 {
		return cmp
	}
	if a.ProcessSteps[0] == "identify-problem" || a.ProcessSteps[0] == "game-analysis" {
		return -1
	}

	return strings.Compare(a.Title, b.Title)
}

func writeNotebookHTMLToFile(notebookName string, frontmatter []RenderedEntry, entries []RenderedEntry, focusList []FocusGroup) {
	f, err := os.Create(path.Join(tmpDir, notebookName+".html"))
	must(err)
	err = f.Truncate(0)
	must(err)
	defer f.Close()

	err = templateFile.Execute(f, Notebook{
		Date:        time.Now(),
		Notebook:    cases.Title(language.AmericanEnglish).String(notebookName),
		Frontmatter: frontmatter,
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
