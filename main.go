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
	EntriesPath       string
	AssetsPath        string
	OutputPath        string
	Port              int
	IncludeUnfinished bool
}

var tmpDir string = "temp"

//go:embed Static/*
var StaticFiles embed.FS

func ParseArgs() Config {
	cfg := Config{
		EntriesPath:       "",
		AssetsPath:        "",
		OutputPath:        "",
		IncludeUnfinished: false,
	}

	flag.StringVar(&cfg.EntriesPath, "entries", "", "OS Path to the entries (following metadata all that stuff)")
	flag.StringVar(&cfg.AssetsPath, "assets", "", "OS Path to the image assets (the directory that all the images are in)")
	flag.StringVar(&cfg.OutputPath, "output", "", "where to save the output PDFS. Leave blank to just serve for ever rather than make PDFs")
	flag.IntVar(&cfg.Port, "port", 32124, "the port to serve on. (Dont make this 0)")
	flag.BoolVar(&cfg.IncludeUnfinished, "includeUnfinished", false, "Include unfinished entries. By default, skip entries that do not have the finished checkbox checked")

	flag.Parse()

	failed := false
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

func main() {
	args := ParseArgs()
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
	makeNotebookFile("hardware", notes, []string{"How this Notebook is Organized"}, args.IncludeUnfinished)
	makeNotebookFile("software", notes, []string{"How this Notebook is Organized"}, args.IncludeUnfinished)
	makeNotebookFile("strategy", notes, []string{"How this Notebook is Organized", "Meet the Team", "Meet the Bears Behind the Bots", "The Engineering Design Process"}, args.IncludeUnfinished)

	log.Println("Made HTML")

	close := startFileServing(tmpDir, args.Port)
	defer close()

	onlyServe := args.OutputPath == ""
	if onlyServe {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("Serving at http://localhost:%d, press ctrl+c to exit\n", args.Port)
		<-done // Will block here until user hits ctrl+c
	} else {
		log.Println("Rendering")
		for _, notebook := range []string{"hardware", "software", "strategy"} {
			url := fmt.Sprintf("http://localhost:%d/%s.html", args.Port, notebook)
			err := savePageToPdf(url, path.Join(args.OutputPath, notebook+".pdf"))
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Finished", notebook)
		}
		log.Println("Finished saving")
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
	ents, err := StaticFiles.ReadDir("Static")
	must(err)
	for _, ent := range ents {
		if ent.IsDir() {
			log.Print("DONT PUT DIRECTORIES IN THE STATIC FILES DIRECTORY (or do just like, update the code to handle them)")
			continue
		}
		bs, err := StaticFiles.ReadFile(path.Join("Static", ent.Name()))
		must(err)
		err = os.WriteFile(path.Join(tmpDir, ent.Name()), bs, 0o644)
		must(err)
	}
}
