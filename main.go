package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"text/template"
)

type Config struct {
	Notebook  string // Hardware, Software, Strategy
	VaultPath string
}

func PrintUsage() {
	usage := `
gen-notebook notebook_type path_to_entries
notebook_type: one of 'software', 'hardware', or 'strategy
`
	fmt.Println(usage)
	os.Exit(1)
}

func ParseArgs() Config {
	args := os.Args
	if len(args) < 3 {
		PrintUsage()
	}
	notebook := os.Args[1]
	notebook = strings.TrimSpace(notebook)
	notebook = strings.ToLower(notebook)
	if !(notebook == "software" || notebook == "hardware" || notebook == "strategy") {
		PrintUsage()
	}
	vault_path := os.Args[2]

	return Config{
		Notebook:  notebook,
		VaultPath: vault_path,
	}
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

type RenderedEntry struct {
	Data Note

	PrevNote, NextNote *string

	Html string
}

type FocusGroup struct {
	Focus   string
	Entries []Note
}

type Notebook struct {
	Notebook string
	Entries  []RenderedEntry
	ByFocus  []FocusGroup
}

func parseFiles(files []string) (mds []Note, errs []error) {
	for _, file := range files {
		m, err := getMetadata(file)
		if errors.Is(err, ErrNotEntry) {
			continue
		} else if err != nil {
			errs = append(errs, err)
		} else {
			mds = append(mds, m)
		}
	}
	return mds, errs
}

//go:embed page.tmpl.html
var templateFileSource string

var templateFile = template.Must(template.New("outputPage").Parse(templateFileSource))

func main() {
	args := ParseArgs()
	files := getAllFilesInDirectory(args.VaultPath)
	notes, errs := parseFiles(files)
	if len(errs) > 0 {
		fmt.Println(errs)
	}

	wanted_entries := filterFilesForThisNotebook(notes, args.Notebook)

	entries := []RenderedEntry{}
	frontmatter := []RenderedEntry{}

	errs = []error{}
	byFocus := map[string][]Note{}

	for _, metadata := range wanted_entries {
		// Sort by focus
		if l, exists := byFocus[metadata.Focus]; exists {
			byFocus[metadata.Focus] = append(l, metadata)
		} else {
			byFocus[metadata.Focus] = []Note{metadata}
		}

		buf := bytes.NewBuffer([]byte{})

		err := NotebookRender().Render(buf, metadata.Src, metadata.Doc)
		if err != nil {
			errs = append(errs, err)
		}
		// Rout to frontmatter or regular entry
		if slices.Contains(metadata.ProcessSteps, "frontmatter") {
			frontmatter = append(frontmatter, RenderedEntry{
				Data: metadata,
				Html: buf.String(),
			})
		} else {
			entries = append(entries, RenderedEntry{
				Data: metadata,
				Html: buf.String(),
			})

		}

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
		return strings.Compare(a.Focus, b.Focus)
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
			// fmt.Println("Found prev", i, byFocus[focus][i].PrevInFocus)
		}
		if i < len(neighbors)-1 {
			entries[j].Data.NextInFocus = &neighbors[i+1]
		}
	}

	writeToFile(args.Notebook, entries, focusList)
}

func writeToFile(notebookName string, entries []RenderedEntry, focusList []FocusGroup) {
	f, err := os.Create("Out/index.html")
	must(err)
	err = f.Truncate(0)
	must(err)
	defer f.Close()

	err = templateFile.Execute(f, Notebook{
		Notebook: notebookName,
		Entries:  entries,
		ByFocus:  focusList,
	})
	must(err)

}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
