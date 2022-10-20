package main

import (
	"log"
	"os"
	"strconv"

	"github.com/alecthomas/kong"
	"github.com/toastate/toastfront/internal/builder"
	"github.com/toastate/toastfront/internal/server"
	"github.com/toastate/toastfront/internal/tlogger"
	"github.com/toastate/toastfront/pkg/config"
)

var CLI struct {
	Build CommandBuild `cmd:"" aliases:"b" help:"Builds or rebuilds the project."`
	Serve CommandServe `cmd:"" aliases:"s" help:"Run a live dev server."`

	ConfigFile string `short:"c" help:"configuration file path (optional)"`
}

type CommandBuild struct {
	SrcDir   string `help:"Source directory." type:"existingdir"`
	BuildDir string `help:"Build output."`

	Verbose int `short:"v" help:"Print verbose output." type:"counter"`
}

type CommandServe struct {
	SrcDir   string `help:"Source directory." type:"existingdir"`
	BuildDir string `help:"Build output."`
	Build    bool   `negatable:"" help:"Don't run build."`

	Port int `short:"p" help:"Listener port"`

	Verbose int `short:"v" help:"Print verbose output." type:"counter"`
}

func main() {
	ctx := kong.Parse(&CLI, kong.UsageOnError())

	err := config.Init(CLI.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err = ctx.Run(ctx)
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

	buildtool := builder.NewBuilder(r.SrcDir, r.BuildDir, ".")

	err := buildtool.Init()
	if err != nil {
		return err
	}

	err = buildtool.Build()
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
	if r.Port <= 0 {
		r.Port = config.Config.ServeConfig.Port
	}

	serv := server.NewServer(r.SrcDir, r.BuildDir, ".", strconv.Itoa(r.Port), config.Config.ServeConfig.Redirect404)

	return serv.Start(!r.Build)
}
