# Toastfront

Toastfront is a simple and transparent front-end build system designed primarily for vanilla JavaScript projects. It is built with a focus on file imports and internationalization support. Toastfront has no dependencies and generates vanilla HTML, CSS, and JS files, which provides full native performance in the browser.

Toastfront differs from frameworks like ReactJS or AngularJS. While React and Angular are comprehensive libraries or frameworks for building complex, interactive web applications with a focus on UI components and data management, Toastfront is primarily focused on vanilla websites and on simplifying the process of importing and bundling HTML, CSS, and JavaScript files.

Toastfront is designed to work mainly with vanilla JavaScript projects but should also work with other projects that don't require transpiling. It does not provide the same level of abstraction, structure, or built-in components as React or Angular. Instead, Toastfront is intended to help developers organize their projects and streamline the build process with features like HTML templating, internationalization, and file imports.

In summary, while ReactJS and AngularJS are full-featured frameworks for creating rich web applications, Toastfront is a lightweight build system aimed at simplifying the process of bundling and organizing front-end assets for projects that do not require the complexity of a full-fledged framework.

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

# Contributing

PRs are always welcome! To get started, join our [Discord](https://discord.gg/NC8sgX6E75).

## Help

Please open an issue if you need help using toastfront

## License 

This project is licensed under the terms of the GNU General Public License v3.0.

