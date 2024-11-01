package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Config struct {
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
	if len(args) < 2 {
		PrintUsage()
	}
	vault_path := os.Args[1]

	return Config{
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
	Date        time.Time
	Notebook    string
	Frontmatter []RenderedEntry
	Entries     []RenderedEntry
	ByFocus     []FocusGroup
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

func makeNotebookFile(notebook string, allNotes []Note, frontmatterWanted []string) {

	wanted_entries := filterFilesForThisNotebook(allNotes, notebook)

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
	slices.SortFunc(entries, func(a, b RenderedEntry) int {
		return a.Data.Date.Compare(b.Data.Date)
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

	for _, fnote := range frontmatterNotes {
		fmt.Println("Frontmatter: ", fnote.Data.Title)
	}

	writeNotebookHTMLToFile(notebook, orderedFrontmatterNotes, entries, focusList)

}

func main() {
	args := ParseArgs()
	files := getAllFilesInDirectory(args.VaultPath)
	notes, errs := parseFiles(files)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println(e)
		}
	}
	makeNotebookFile("hardware", notes, []string{"How this Notebook is Organized"})
	makeNotebookFile("software", notes, []string{"How this Notebook is Organized"})
	makeNotebookFile("strategy", notes, []string{"How this Notebook is Organized", "Meet the Team", "Meet the Bears Behind the Bots", "The Engineering Design Process"})

	log.Println("Made HTML")
	port := 8080

	close := startFileServing("Out/", port)
	defer close()
	time.Sleep(1 * time.Hour)
	//
	// for _, notebook := range []string{"hardware", "software", "strategy"} {
	// url := fmt.Sprintf("http://localhost:%d/%s.html", port, notebook)
	// err := savePageToPdf(url, "PDFs/"+notebook+".pdf")
	// if err != nil {
	// log.Fatal(err)
	// }
	// log.Println("Finished ", notebook)
	// }
	// log.Println("Finished saving")

}

func writeNotebookHTMLToFile(notebookName string, frontmatter []RenderedEntry, entries []RenderedEntry, focusList []FocusGroup) {
	f, err := os.Create(fmt.Sprintf("Out/%s.html", notebookName))
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

func must(e error) {
	if e != nil {
		panic(e)
	}
}
