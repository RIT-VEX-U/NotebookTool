package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MetadataError struct {
	file string
	err  error
}

func (me MetadataError) Unwrap() error {
	return me.err
}

func (me MetadataError) Error() string {
	return "in " + me.file + ": " + me.err.Error()
}

type Note struct {
	Location string

	Title string
	Focus string

	Notebook     string
	ProcessSteps []string
	Date         time.Time
	Robots       []string

	Authors      []string
	Proofreaders []string
	Doc          ast.Node
	Src          []byte

	PrevInFocus *Note
	NextInFocus *Note
}

func (n Note) LucideIcon() string {
	icon := "menu"
	strokeColor := "black"

	switch n.ProcessSteps[0] {
	case "other":
		icon = "circle-ellipsis"
		strokeColor = "var(--proc-step-other)"
	case "identify-problem":
		icon = "scan-eye"
		strokeColor = "var(--proc-step-identify)"
	case "update":
		icon = "refresh-cw"
		strokeColor = "var(--proc-step-update)"
	case "brainstorm":
		icon = "brain"
		strokeColor = "var(--proc-step-brainstorm)"
	case "select-best":
		icon = "mouse-pointer-click"
		strokeColor = "var(--proc-step-select)"
	case "test":
		icon = "clipboard-list"
		strokeColor = "var(--proc-step-test)"
	case "context":
		icon = "zoom-in"
		strokeColor = "var(--proc-step-context)"
	case "game-analysis":
		icon = "chart-pie"
		strokeColor = "var(--proc-step-game-analysis)"
	}

	return fmt.Sprintf("<i data-lucide=\"%s\" style = \"stroke: %s;\"></i>", icon, strokeColor)
}

var dateRemover = regexp.MustCompile(`^\d{1,2}-\d{1,2}-\d{2,4}\s+`)

func (m Note) Anchor() string {

	name := strings.ReplaceAll(strings.ToLower(m.Title), " ", "-")
	name = dateRemover.ReplaceAllString(name, "")
	date := m.Date.Format("01-02-2006")
	return fmt.Sprintf("entry-%s-%s", name, date)
}

func (m Note) String() string {
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

var ErrNotEntry error = errors.New("not a notebook entry")

func extractMetadata(meta map[string]interface{}) (Note, error) {
	out := Note{}
	nbI, exists := meta["notebook"]
	if !exists {
		return out, ErrNotEntry
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
	out.Robots, _ = getStringlist("robot", meta)
	if len(out.ProcessSteps) == 0 {
		return out, fmt.Errorf("missing process_step")
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

	newTitle := dateRemover.ReplaceAllString(title, "")
	return strings.TrimSpace(string(newTitle))
}

func getMetadata(filepath string) (Note, error) {
	bs, err := os.ReadFile(filepath)
	if errors.Is(err, ErrNotEntry) {
		return Note{}, err
	} else if err != nil {
		return Note{}, MetadataError{filepath, err}
	}

	doc := NotebookParser().Parse(text.NewReader(bs))
	meta, err := extractMetadata(doc.OwnerDocument().Meta())
	if err != nil {
		return meta, MetadataError{filepath, err}
	}
	meta.Location = filepath
	meta.Src = bs
	meta.Doc = doc
	meta.Title = fixTitle(filepath)

	dir := path.Dir(filepath)
	meta.Focus = path.Base(dir)

	return meta, nil
}

func filterFilesForThisNotebook(allnotes []Note, notebook string) []Note {
	filtered := []Note{}
	for _, n := range allnotes {
		if n.Notebook == notebook {
			filtered = append(filtered, n)
		}
	}
	return filtered
}
