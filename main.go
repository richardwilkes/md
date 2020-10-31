package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mdigger/goldmark-attributes"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func main() {
	cmdline.AppName = "MarkDown"
	cmdline.AppCmdName = "md"
	cmdline.License = "Mozilla Public License, version 2.0"
	cmdline.CopyrightYears = "2020"
	cmdline.CopyrightHolder = "Richard A. Wilkes"
	cl := cmdline.New(true)
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
		goldmark.WithExtensions(
			extension.GFM,
			extension.NewTypographer(),
			extension.Footnote,
			attributes.Extension,
		),
	)
	for _, p := range cl.Parse(os.Args[1:]) {
		if filepath.Ext(p) != ".md" {
			jot.Warn("skipping non-markdown file: " + p)
			continue
		}
		data, err := ioutil.ReadFile(p)
		jot.FatalIfErr(err)
		var f *os.File
		f, err = os.Create(fs.TrimExtension(p) + ".html")
		jot.FatalIfErr(err)
		jot.FatalIfErr(md.Convert(data, f))
		jot.FatalIfErr(f.Close())
	}
	atexit.Exit(0)
}
