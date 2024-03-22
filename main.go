package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/richardwilkes/md/md"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/xio/fs"
)

func main() {
	cmdline.AppName = "MarkDown"
	cmdline.AppCmdName = "md"
	cmdline.License = "Mozilla Public License, version 2.0"
	cmdline.CopyrightStartYear = "2020"
	cmdline.CopyrightHolder = "Richard A. Wilkes"
	cl := cmdline.New(true)
	for _, p := range cl.Parse(os.Args[1:]) {
		if filepath.Ext(p) != ".md" {
			slog.Warn("skipping non-markdown file: " + p)
			continue
		}
		data, err := md.MarkdownToHTML(p)
		fatal.IfErr(err)
		fatal.IfErr(os.WriteFile(fs.TrimExtension(p)+".html", data, 0o644))
	}
	atexit.Exit(0)
}
