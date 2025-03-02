package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

type Config struct {
	EntriesPath string
	AssetsPath  string
	OutputPath  string

	// html file that will be pasted onto the front of the PDF
	FrontPagePath string

	// port to serve on
	Port              int
	IncludeUnfinished bool

	// When creating the notebook
	MakeTemplatePath string
}

var tmpDir string = "temp"

//go:embed static/*
var StaticFiles embed.FS

func ParseArgs() Config {
	cfg := Config{
		EntriesPath:       "",
		AssetsPath:        "",
		OutputPath:        "",
		IncludeUnfinished: false,
		MakeTemplatePath:  "",
	}

	flag.StringVar(&cfg.EntriesPath, "entries", "", "OS Path to the entries (following metadata all that stuff)")
	flag.StringVar(&cfg.AssetsPath, "assets", "", "OS Path to the image assets (the directory that all the images are in)")
	flag.StringVar(&cfg.OutputPath, "output", "", "where to save the output PDFS. Leave blank to just serve for ever rather than make PDFs")
	flag.IntVar(&cfg.Port, "port", 32124, "the port to serve on. (Dont make this 0)")
	flag.BoolVar(&cfg.IncludeUnfinished, "includeUnfinished", false, "Include unfinished entries. By default, skip entries that do not have the finished checkbox checked")

	flag.StringVar(&cfg.MakeTemplatePath, "make-template", "", "directory to place a template notebook for you to fill in")

	flag.StringVar(&cfg.FrontPagePath, "front-page", "", "Path to html file that will be included as the first page")

	flag.Parse()

	failed := false
	if cfg.MakeTemplatePath != "" {
		// don't care about the rest of the args
		return cfg
	}
	if cfg.EntriesPath == "" || cfg.AssetsPath == "" || cfg.Port == 0 {
		failed = true
	}

	if failed {
		flag.Usage()
		os.Exit(1)
	}
	return cfg

}

type RenderedEntry struct {
	Data               Note
	PrevNote, NextNote *string
	Html               string
	GradientHR         string
}

type FocusGroup struct {
	Notebook string
	Focus    string
	Entries  []Note
}

type Notebook struct {
	Date        time.Time
	Notebook    string
	FrontPage   string
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

//go:embed NotebookTemplate
var notebook_template embed.FS

func MakeTemplateAtPath(directory string) {
	if _, err := os.Stat(directory); !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("There is already something where you're trying to create notebook")
	}

	err := os.CopyFS(directory, notebook_template)
	if err != nil {
		log.Fatalf("Failed to make the notebook template: %v", err)
	}
	err = os.Rename(path.Join(directory, "NotebookTemplate"), path.Join(directory, "Notebook"))
	if err != nil {
		log.Fatalf("Failed to rename notebook template folder: %v", err)
	}
}

func main() {
	args := ParseArgs()

	if args.MakeTemplatePath != "" {
		log.Println("Creating notebook template at ", args.MakeTemplatePath)
		MakeTemplateAtPath(args.MakeTemplatePath)
		return
	}

	files := getAllFilesInDirectory(args.EntriesPath)
	notes, errs := parseFiles(files)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println(e)
		}
	}

	log.Println("Setting up")
	setupTmpOutputDir(args)

	log.Println("Making notebooks")

	makeNotebookFile(notes, []string{"How this Notebook is Organized", "Meet the Team", "Meet the Bears Behind the Bots", "The Engineering Design Process"}, args.IncludeUnfinished, args.FrontPagePath, "notebook.html")

	log.Println("Made HTML")

	stopServer := startFileServing(tmpDir, args.Port)
	defer stopServer()

	onlyServe := args.OutputPath == ""
	if onlyServe {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("Serving at http://localhost:%d/notebook.html, press ctrl+c to exit\n", args.Port)
		<-done // Will block here until user hits ctrl+c
	} else {
		log.Fatalf("Notebook auto printing is broken bc chrome dies at big PDFs. Instead, omit the -output argument and open the link it gives you, press Ctrl-P, and print to PDF")
	}

}

func must(e error) {
	if e != nil {
		panic(e)
	}
}

func setupTmpOutputDir(cfg Config) {
	os.RemoveAll(tmpDir)
	err := os.MkdirAll(tmpDir, 0o755)
	must(err)
	err = os.Symlink(cfg.AssetsPath, path.Join(tmpDir, "Assets"))
	must(err)
	ents, err := StaticFiles.ReadDir("static")
	must(err)
	for _, ent := range ents {
		if ent.IsDir() {
			log.Print("DONT PUT DIRECTORIES IN THE STATIC FILES DIRECTORY (or do just like, update the code to handle them)")
			continue
		}
		bs, err := StaticFiles.ReadFile(path.Join("static", ent.Name()))
		must(err)
		err = os.WriteFile(path.Join(tmpDir, ent.Name()), bs, 0o644)
		must(err)
	}
}
