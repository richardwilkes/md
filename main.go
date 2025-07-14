package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/richardwilkes/md/md"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xflag"
	"github.com/richardwilkes/toolbox/v2/xos"
)

func main() {
	xos.AppName = "MarkDown"
	xos.AppCmdName = "md"
	xos.License = "Mozilla Public License, version 2.0"
	xos.CopyrightStartYear = "2020"
	xos.CopyrightHolder = "Richard A. Wilkes"
	xos.AppVersion = "1.0"
	xflag.AddVersionFlags()
	xflag.SetUsage(nil, "", "")
	xflag.Parse()
	for _, p := range flag.Args() {
		if filepath.Ext(p) != ".md" {
			slog.Warn("skipping non-markdown file: " + p)
			continue
		}
		data, err := md.MarkdownToHTML(p)
		xos.ExitIfErr(err)
		xos.ExitIfErr(os.WriteFile(xfilepath.TrimExtension(p)+".html", data, 0o644))
	}
	xos.Exit(0)
}
