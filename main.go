package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/richardwilkes/md/md"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
)

func main() {
	cmdline.AppName = "MarkDown"
	cmdline.AppCmdName = "md"
	cmdline.License = "Mozilla Public License, version 2.0"
	cmdline.CopyrightYears = "2020"
	cmdline.CopyrightHolder = "Richard A. Wilkes"
	cl := cmdline.New(true)
	for _, p := range cl.Parse(os.Args[1:]) {
		if filepath.Ext(p) != ".md" {
			jot.Warn("skipping non-markdown file: " + p)
			continue
		}
		data, err := md.MarkdownToHTML(p)
		jot.FatalIfErr(err)
		jot.FatalIfErr(ioutil.WriteFile(fs.TrimExtension(p)+".html", data, 0644))
	}
	atexit.Exit(0)
}
