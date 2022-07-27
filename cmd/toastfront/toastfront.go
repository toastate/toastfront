package main

import (
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/toastate/toastfront/internal/builder"
	"github.com/toastate/toastfront/internal/server"
	"github.com/toastate/toastfront/internal/tlogger"
	"github.com/toastate/toastfront/internal/watcher"
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
	if r.BuildDir == "" {
		r.BuildDir = "build"
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

	if r.SrcDir == "" {
		r.SrcDir = "src"
	}
	if r.BuildDir == "" {
		r.BuildDir = "build"
	}

	buildtool := builder.Builder{
		SrcDir:     r.SrcDir,
		BuildDir:   r.BuildDir,
		RootFolder: ".",
	}

	buildStart := time.Now()
	err := buildtool.Build()
	estBuildTime := time.Now().Sub(buildStart)
	estBuildTime *= 2
	if estBuildTime > time.Millisecond*500 {
		estBuildTime = time.Millisecond * 500
	}

	if err != nil {
		os.Exit(1)
	}

	updates := watcher.StartWatcher(r.SrcDir)

	go func() {
		for {
			_ = <-updates
		rootFor:
			for {
				select {
				case <-updates:
					continue
				case <-time.After(time.Millisecond * 500):
					break rootFor
				}
			}
			buildtool.Build()
			server.TriggerReload()
		}
	}()
	// Start file change listener
	// Builder: add dep tree
	// Check dependancy tree on update & rebuild
	// k: Start server

	server.Start(r.BuildDir, "8100")

	return nil
}
