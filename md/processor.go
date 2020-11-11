package md

import (
	"bytes"
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	attributes "github.com/mdigger/goldmark-attributes"
	replacer "github.com/mdigger/goldmark-text-replacer"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

var (
	includeDirective      = []byte(":include:")
	includeRegexDirective = []byte(":include*:")
	cssDirective          = []byte(":css:")
	titleDirective        = []byte(":title:")
)

// MarkdownToHTML converts the specified markdown file into HTML, processing it for include, css, and title directives
// prior to processing it for markdown.
func MarkdownToHTML(file string) ([]byte, error) {
	doc := &processor{
		cssMap: make(map[string]bool),
	}
	if err := doc.include(file); err != nil {
		return nil, err
	}
	return doc.markdownToHTML()
}

type processor struct {
	css    []string
	cssMap map[string]bool
	title  string
	data   bytes.Buffer
	depth  int
}

func (p *processor) include(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errs.Wrap(err)
	}
	path = filepath.Dir(path)
	for len(data) > 0 {
		var line []byte
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			line = data[:i]
			data = data[i+1:]
		} else {
			line = data
			data = nil
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		switch {
		case bytes.HasPrefix(line, includeDirective):
			p.depth++
			if err = p.include(filepath.Join(path, string(line[len(includeDirective):]))); err != nil {
				return err
			}
			p.depth--
		case bytes.HasPrefix(line, includeRegexDirective):
			p.depth++
			parts := strings.SplitN(string(line[len(includeRegexDirective):]), "|", 2)
			if len(parts) != 2 {
				return errs.Newf("invalid %s directive: %s", includeRegexDirective, line)
			}
			if err = p.includeRegex(filepath.Join(path, parts[0]), parts[1]); err != nil {
				return err
			}
			p.depth--
		case bytes.HasPrefix(line, cssDirective):
			css := filepath.Join(path, string(line[len(cssDirective):]))
			if !p.cssMap[css] {
				p.cssMap[css] = true
				p.css = append(p.css, css)
			}
		case bytes.HasPrefix(line, titleDirective):
			if p.depth == 0 || p.title == "" {
				p.title = string(line[len(titleDirective):])
			}
		default:
			p.data.Write(line)
			p.data.WriteByte('\n')
		}
	}
	return nil
}

func (p *processor) includeRegex(dir, pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return errs.Wrap(err)
	}
	var f *os.File
	if f, err = os.Open(dir); err != nil {
		return errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(f)
	var fis []os.FileInfo
	if fis, err = f.Readdir(0); err != nil {
		return errs.Wrap(err)
	}
	names := make([]string, 0, len(fis))
	for _, fi := range fis {
		if !fi.IsDir() {
			name := fi.Name()
			if strings.HasSuffix(strings.ToLower(name), ".md") && regex.MatchString(name) {
				names = append(names, name)
			}
		}
	}
	txt.SortStringsNaturalAscending(names)
	for _, name := range names {
		if err = p.include(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

func (p *processor) markdownToHTML() ([]byte, error) {
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.GFM,
			extension.NewTypographer(),
			extension.Footnote,
			attributes.Extension,
			replacer.New(
				"^1^", "&sup1;",
				"^2^", "&sup2;",
				"^3^", "&sup3;",
				"!1/2!", "&frac12;",
				"!1/3!", "&frac13;",
				"!1/4!", "&frac14;",
				"!1/5!", "&frac15;",
				"!1/6!", "&frac16;",
				"!1/8!", "&frac18;",
				"!2/3!", "&frac23;",
				"!2/5!", "&frac25;",
				"!3/4!", "&frac34;",
				"!3/5!", "&frac35;",
				"!3/8!", "&frac38;",
				"!4/5!", "&frac45;",
				"!5/6!", "&frac56;",
				"!5/8!", "&frac58;",
				"!7/8!", "&frac78;",
			),
		),
	)
	var buffer bytes.Buffer
	buffer.WriteString(`<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta http-equiv="x-ua-compatible" content="ie=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>`)
	buffer.WriteString(html.EscapeString(p.title))
	buffer.WriteString("</title>\n")
	for _, css := range p.css {
		buffer.WriteString(`	<link rel="stylesheet" type="text/css" href="`)
		buffer.WriteString(css)
		buffer.WriteString("\">\n")
	}
	buffer.WriteString("</head>\n<body>\n")
	if err := md.Convert(p.data.Bytes(), &buffer); err != nil {
		return nil, errs.Wrap(err)
	}
	buffer.WriteString("</body>\n</html>\n")
	return buffer.Bytes(), nil
}
