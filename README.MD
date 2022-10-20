# Toastfront

The simple front-end build tool used by the [Toastate.com](https://www.toastate.com) team 

## Description

Toastfront is a simple front-end build system relying on file imports and with internationalisation out of the box.

Toastfront is made to be used mostly with vanilla javascript in mind but should work with every project that doesn't need transpiling.

It is transparent because it has no dependencies and outputs vanilla HTML/css/js. It provides code with full native performance in the browser.

## Getting started - Command Line Interface

### Installation

Grab the executable that matches your OS in the release tab

Or clone it and build it from scratch using `go build cmd/toastfront/toastfront.go`

### Setup your project

Either start from scratch or by copying the `exemple` folder

Create a `toastatfront.json` file at the root of your project [toastfront.json documentation](https://github.com/toastate/toastfront/wiki/toastfront.json-structure)

### Take advantage of toastfront features

The complete [list of features is avaliable on the wiki](https://github.com/toastate/toastfront/wiki/Toastfront-features) including:
    - HTML Templating & Internationalisation (via vars)
    - HTML Imports
    - JS Imports
        - Importing HTML vars in your file
    - CSS Imports

### Run it

Use `toastfront serve` to start a live development server (Accessible by default via http://localhost:8100)

Use `toastfront build` to create a production ready deployement of your project (avaliable by default in the build/ folder)


## Getting started - Golang Package

### Installation

```
go get -u github.com/toastate/toastfront
```

### Quickstart

Add these import lines to the file you are working in:

```go
import "github.com/toastate/toastfront/pkg/config"
import "github.com/toastate/toastfront/pkg/server"
import "github.com/toastate/toastfront/pkg/builder"
```

To setup or modify the default configuration:
```go
err := config.Init("/path/to/configuration/file")
if err != nil {
    // handle error
}

config.Config.BuildDir = "anotherName"
```

To build a toastfront directory:
```go
bd := builder.NewBuilder("source directory", "target directory", ".")

err := bd.Init()
if err != nil {
    // handle error
}

err := bd.Build()
if err != nil {
    // handle error
}
```

To serve a toastfront directory and live reload it on file changes:
```go
serv := server.NewServer("source directory", "target directory", ".", "8100", "")

err := serv.Start(true)
if err != nil {
    // handle error
}
```

## Help

Please open an issue if you need help using toastfront

## License 

This project is licensed under the terms of the GNU General Public License v3.0.

