package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
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

func getAllFiles(root string) []string {
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

func main() {
	args := ParseArgs()
	// log.Println(args)
	files := getAllFiles(args.VaultPath)
	// log.Println(files)
	mds := []Note{}
	errs := []error{}

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
	if len(errs) > 0 {
		fmt.Println(errs)
	}

	wanted_entries := listOfFilesInThisNotebook(mds, args.Notebook)

	// fmt.Println(wanted)

	f, err := os.OpenFile("Out/index.html", os.O_WRONLY|os.O_CREATE, 0666)
	must(err)
	err = f.Truncate(0)
	must(err)
	defer f.Close()

	t, err := template.ParseFiles("page.tmpl.html")
	must(err)

	entries := []RenderedEntry{}

	errs = []error{}
	byFocus := map[string][]Note{}

	for _, metadata := range wanted_entries {
		if l, exists := byFocus[metadata.Focus]; exists {
			byFocus[metadata.Focus] = append(l, metadata)
		} else {
			byFocus[metadata.Focus] = []Note{metadata}
		}

		buf := bytes.NewBuffer([]byte{})

		err = NotebookRender().Render(buf, metadata.Src, metadata.Doc)
		if err != nil {
			errs = append(errs, err)
		}
		entries = append(entries, RenderedEntry{
			Data: metadata,
			Html: buf.String(),
		})
	}
	if len(errs) > 0 {
		fmt.Println(errs)
	}

	focusList := []FocusGroup{}
	for focus, entries := range byFocus {
		focusList = append(focusList, FocusGroup{
			Focus:   focus,
			Entries: entries,
		})
	}

	err = t.Execute(f, Notebook{
		Notebook: args.Notebook,
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
