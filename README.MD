# Toastfront

The simple front-end build used by the [Toastate.com](https://www.toastate.com) team 

## Description

Toastfront was designed to be a simple front-end build system relying on file imports and with internationalisation out of the box

Toastfront is made to be used mostly with vanilla javascript in mind but should work with every project that doesn't need transpiling

## Getting started

#### Installing

Grab the executable that matches your OS in the release tab

Or clone it and build it from scratch using `go build cmd/toastfront/toastfront.go`

#### Setup your project

Either start from scratch or by copying the `exemple` folder

Create a `toastatfront.json` file at the root of your project
[[Here link to toastfront json wiki page]]

#### Take advantage of toastfront features

The complete list of features is avaliable on the wiki including:
    - I18N setup
    - HTML Imports
    - HTML Templating & Internationalisation
    - JS Imports
        - Importing vars in your file
    - CSS Imports

#### Run it

Use `toastfront serve` to start a live development server (Accessible by default via http://localhost:8100)

Use `toastfront build` to create a production ready deployement of your project (avaliable by default in the build/ folder)


## Help

Please open an issue if you need help using toastfront

## License 

This project is licensed under the terms of the GNU General Public License v3.0.

