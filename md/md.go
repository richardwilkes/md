package md

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio"
)

const (
	includeDirective = ":include:"
	cssDirective     = ":css:"
	titleDirective   = ":title:"
	idDirective      = ":id:"
	idAttribute      = "id"
	styleDirective   = ":style:"
	styleAttribute   = "style"
	classDirective   = ":class:"
	classAttribute   = "class"
)

var (
	headerPrefix = regexp.MustCompile(`^#+\s`)
)

// MarkDown provide MarkDown to HTML processing.
type MarkDown struct {
	includeProvider func(path string) (io.ReadCloser, error)
	lineBufferSize  int
	depth           int
	lines           []string
	lineDirectives  []map[string]string
	lineDirective   map[string]string
	css             []string
	title           string
}

// New creates a new MarkDown processor with the specified options.
func New(options ...Option) (*MarkDown, error) {
	m := &MarkDown{
		includeProvider: func(path string) (io.ReadCloser, error) { return os.Open(path) },
		lineBufferSize:  bufio.MaxScanTokenSize,
	}
	for _, option := range options {
		if err := option(m); err != nil {
			return nil, errs.Wrap(err)
		}
	}
	return m, nil
}

// ConvertFileToHTML reads the MarkDown from the given path, converts it to HTML, and writes it to the given path.
func (m *MarkDown) ConvertFileToHTML(in, out string) (err error) {
	var inFile *os.File
	inFile, err = os.Open(in)
	if err != nil {
		return errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(inFile)
	var outFile *os.File
	if outFile, err = os.Create(out); err != nil {
		return errs.Wrap(err)
	}
	defer func() {
		if closeErr := outFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	return m.ConvertToHTML(filepath.Dir(in), inFile, outFile)
}

// ConvertToHTML converts the MarkDown from the input and writes an HTML version to the output.
func (m *MarkDown) ConvertToHTML(dir string, in io.Reader, w io.Writer) error {
	m.reset()
	defer m.reset()
	if err := m.processIntoLines(dir, in); err != nil {
		return err
	}
	bw := bufio.NewWriter(w)
	if err := m.writeHTMLPrefix(bw); err != nil {
		return err
	}
	for lineNum, line := range m.lines {
		switch {
		case headerPrefix.MatchString(line) && (lineNum == 0 || m.lines[lineNum-1] == ""):
			i := 1
			for line[i] == '#' {
				i++
			}
			line = strings.TrimSpace(line[i:])
			if _, err := fmt.Fprintf(bw, "<h%d", i); err != nil {
				return errs.Wrap(err)
			}
			if err := m.emitHTMLAttribute(bw, idAttribute, m.lineDirectives[lineNum][idDirective]); err != nil {
				return errs.Wrap(err)
			}
			if err := m.emitHTMLAttribute(bw, classAttribute, m.lineDirectives[lineNum][classDirective]); err != nil {
				return errs.Wrap(err)
			}
			if err := m.emitHTMLAttribute(bw, styleAttribute, m.lineDirectives[lineNum][styleDirective]); err != nil {
				return errs.Wrap(err)
			}
			if _, err := fmt.Fprintf(bw, ">%s</h%d>\n", line, i); err != nil {
				return errs.Wrap(err)
			}
		default:
			if _, err := fmt.Fprintln(bw, line); err != nil {
				return errs.Wrap(err)
			}
		}
	}
	if err := m.writeHTMLPostfix(bw); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func (m *MarkDown) reset() {
	m.depth = 0
	m.lines = nil
	m.lineDirectives = nil
	m.lineDirective = make(map[string]string)
	m.css = nil
}

func (m *MarkDown) processIntoLines(dir string, in io.Reader) (err error) {
	defer errs.Recovery(func(e error) { err = e })
	r := bufio.NewScanner(in)
	r.Buffer(make([]byte, 0, m.lineBufferSize), m.lineBufferSize)
	for r.Scan() {
		line := r.Text()
		switch {
		case strings.HasPrefix(line, includeDirective):
			m.depth++
			if err = m.include(filepath.Join(dir, line[len(includeDirective):])); err != nil {
				return err
			}
			m.depth--
		case strings.HasPrefix(line, cssDirective):
			m.css = append(m.css, filepath.Join(dir, line[len(cssDirective):]))
		case strings.HasPrefix(line, titleDirective):
			if m.depth == 0 || m.title == "" {
				m.title = line[len(titleDirective):]
			}
		case strings.HasPrefix(line, idDirective):
			m.lineDirective[idDirective] = line[len(idDirective):]
		case strings.HasPrefix(line, styleDirective):
			if style, ok := m.lineDirective[styleDirective]; ok {
				m.lineDirective[styleDirective] = style + "; " + line[len(styleDirective):]
			} else {
				m.lineDirective[styleDirective] = line[len(styleDirective):]
			}
		case strings.HasPrefix(line, classDirective):
			if class, ok := m.lineDirective[classDirective]; ok {
				m.lineDirective[classDirective] = class + " " + line[len(classDirective):]
			} else {
				m.lineDirective[classDirective] = line[len(classDirective):]
			}
		case strings.TrimSpace(line) == "":
			m.lines = append(m.lines, "")
			m.lineDirectives = append(m.lineDirectives, nil)
			m.lineDirective = make(map[string]string)
		default:
			m.lines = append(m.lines, line)
			m.lineDirectives = append(m.lineDirectives, m.lineDirective)
			m.lineDirective = make(map[string]string)
		}
	}
	if err = r.Err(); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func (m *MarkDown) include(path string) error {
	f, err := m.includeProvider(path)
	if err != nil {
		return errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(f)
	return m.processIntoLines(filepath.Dir(path), f)
}

func (m *MarkDown) writeHTMLPrefix(w *bufio.Writer) error {
	if _, err := fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta http-equiv="x-ua-compatible" content="ie=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
`, html.EscapeString(m.title)); err != nil {
		return errs.Wrap(err)
	}
	for _, css := range m.css {
		if _, err := fmt.Fprintf(w, `	<link rel="stylesheet" type="text/css" href="%s">
`, css); err != nil {
			return errs.Wrap(err)
		}
	}
	if _, err := w.WriteString(`</head>
<body>
`); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func (m *MarkDown) writeHTMLPostfix(w *bufio.Writer) error {
	if _, err := w.WriteString(`</body>
</html>
`); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func (m *MarkDown) emitHTMLAttribute(w *bufio.Writer, name, value string) error {
	if value != "" {
		if _, err := fmt.Fprintf(w, ` %s="%s"`, name, value); err != nil {
			return errs.Wrap(err)
		}
	}
	return nil
}
