package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MetadataError struct {
	file string
	err  error
}

func (me MetadataError) Error() string {
	return "in " + me.file + ": " + me.err.Error()
}

type Metadata struct {
	Title string
	Focus string

	Notebook     string
	ProcessSteps []string
	Date         time.Time

	Authors      []string
	Proofreaders []string
	Doc          ast.Node
	Src          []byte
}

func (m Metadata) String() string {
	return fmt.Sprintf("Metadata for %v", m.Title)
}

func getStringlist(field string, meta map[string]interface{}) ([]string, error) {
	fieldI, ok := meta[field]
	if !ok {
		return nil, fmt.Errorf("no '%v' field", field)
	}
	if fieldI == nil {
		// empty list
		return []string{}, nil
	}
	f, ok := fieldI.([]interface{})
	if !ok {
		return nil, fmt.Errorf("`%v` field not a list", field)
	}
	res := []string{}
	for _, i := range f {
		if item, ok := i.(string); !ok {
			return res, fmt.Errorf("%v of list field %v not a string. Its a %T", i, field, i)
		} else if item != "" {

			res = append(res, item)
		}
	}
	return res, nil
}

func extractMetadata(meta map[string]interface{}) (Metadata, error) {
	out := Metadata{}
	nbI, exists := meta["notebook"]
	if !exists {
		return out, errors.New("no Notebook Field")
	}
	nb, ok := nbI.(string)
	if !ok {
		return out, errors.New("notebook field of wrong type")
	}
	out.Notebook = nb
	var err error
	if out.Authors, err = getStringlist("authors", meta); err != nil {
		return out, err
	}
	if out.Proofreaders, err = getStringlist("proofread_by", meta); err != nil {
		return out, err
	}
	if out.ProcessSteps, err = getStringlist("process_step", meta); err != nil {
		return out, err
	}

	if dateS, exists := meta["entry_date"]; exists {
		dateS, ok := dateS.(string)
		if !ok {
			return out, fmt.Errorf("entry_date field not the right type")
		}

		t, err := time.Parse("2006-01-02", dateS)
		if err != nil {
			return out, err
		}
		out.Date = t
	} else {
		return out, fmt.Errorf("no `entry_date` field")
	}
	return out, nil
}
func fixTitle(filepath string) string {
	title := path.Base(filepath)
	title = strings.ReplaceAll(title, path.Ext(title), "")
	// regexp.(``)
	r, e := regexp.Compile(`^\d{1,2}-\d{1,2}-\d{2}\s+`)
	if e != nil {
		return title
	}
	newTitle := r.ReplaceAll([]byte(title), []byte{})
	return string(newTitle)
}

func getMetadata(filepath string) (Metadata, error) {
	bs, err := os.ReadFile(filepath)
	if err != nil {
		return Metadata{}, MetadataError{filepath, err}
	}

	doc := NotebookParser().Parse(text.NewReader(bs))
	meta, err := extractMetadata(doc.OwnerDocument().Meta())
	if err != nil {
		return meta, MetadataError{filepath, err}
	}
	meta.Src = bs
	meta.Doc = doc
	meta.Title = fixTitle(filepath)
	dir := path.Dir(filepath)
	meta.Focus = path.Base(dir)

	return meta, nil
}

func listOfFilesInThisNotebook(notes []Metadata, notebook string) []Metadata {
	slices.SortFunc(notes, func(a, b Metadata) int {
		return a.Date.Compare(b.Date)
	})
	filtered := []Metadata{}
	for _, n := range notes {
		if n.Notebook == notebook {
			filtered = append(filtered, n)
		}
	}
	return filtered
}
