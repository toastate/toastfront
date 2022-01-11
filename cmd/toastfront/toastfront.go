package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/toastate/toastfront/internal/builder"
	"github.com/toastate/toastfront/internal/tlogger"
)

var CLI struct {
	Build CommandBuild `cmd:"" help:"Builds or rebuilds the project."`
	Serve CommandServe `cmd:"" help:"Run a live dev server."`
}

type CommandBuild struct {
	SrcDir   string `help:"Source directory." type:"existingdir"`
	BuildDir string `help:"Build output."`

	Verbose int `short:"v" help:"Print verbose output." type:"counter"`
}

type CommandServe struct {
	SrcDir     string `help:"Source directory." type:"existingdir"`
	BuildDir   string `help:"Build output."`
	Build      bool   `negatable:"" help:"Don't run build."`
	LiveReload bool   `negatable:"" help:"Don't watch for changes."`

	Verbose int `short:"v" help:"Print verbose output." type:"counter"`
}

func main() {
	ctx := kong.Parse(&CLI, kong.UsageOnError())

	err := ctx.Run(ctx)

	if err != nil {
		ctx.PrintUsage(false)
	}
}

func applyVerbose(v int) {
	switch v {
	case 0:
		tlogger.ApplyLogLevel("info")
	case 1:
		tlogger.ApplyLogLevel("debug")
	default:
		tlogger.ApplyLogLevel("all")
	}
}

func (r *CommandBuild) Run(ctx *kong.Context) error {
	applyVerbose(r.Verbose)

	if r.SrcDir == "" {
		r.SrcDir = "src"
	}
	if r.SrcDir == "" {
		r.SrcDir = "build"
	}

	buildtool := builder.Builder{
		SrcDir:     r.SrcDir,
		BuildDir:   r.BuildDir,
		RootFolder: ".",
	}

	err := buildtool.Build()

	if err != nil {
		os.Exit(1)
	}

	return nil
}

func (r *CommandServe) Run(ctx *kong.Context) error {
	applyVerbose(r.Verbose)

	return nil
}
