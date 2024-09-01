package main

import (
	"bytes"
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

type Entry struct {
	Data Metadata
	Html string
}

type Notebook struct {
	Notebook string
	Entries  []Entry
}

func main() {
	args := ParseArgs()
	// log.Println(args)
	files := getAllFiles(args.VaultPath)
	// log.Println(files)
	mds := []Metadata{}
	errs := []error{}
	for _, file := range files {
		m, err := getMetadata(file)
		if err != nil {
			errs = append(errs, err)
		} else {
			mds = append(mds, m)
		}
	}
	if len(errs) > 0 {
		fmt.Println(errs)
	}

	wanted := listOfFilesInThisNotebook(mds, args.Notebook)

	// fmt.Println(wanted)

	f, err := os.OpenFile("Out/index.html", os.O_WRONLY|os.O_CREATE, 0666)
	must(err)
	defer f.Close()

	t, err := template.ParseFiles("page.tmpl.html")
	must(err)

	entries := []Entry{}

	errs = []error{}
	for _, w := range wanted {
		buf := bytes.NewBuffer([]byte{})

		err = NotebookRender().Render(buf, w.Src, w.Doc)
		if err != nil {
			errs = append(errs, err)
		}
		entries = append(entries, Entry{
			Data: w,
			Html: buf.String(),
		})
	}
	if len(errs) > 0 {
		fmt.Println(errs)
	}
	t.Execute(f, Notebook{
		Notebook: args.Notebook,
		Entries:  entries,
	})

}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
