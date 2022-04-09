[![Version](https://img.shields.io/badge/goversion-1.18.x-blue.svg)](https://golang.org)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/tsawler/goblender/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/tsawler/toolbox)](https://goreportcard.com/report/github.com/tsawler/toolbox)
![Tests](https://github.com/tsawler/toolbox/actions/workflows/tests.yml/badge.svg)
<a href="https://pkg.go.dev/github.com/tsawler/toolbox"><img src="https://img.shields.io/badge/godoc-reference-%23007d9c.svg"></a>
# Toolbox

A simple example of how to create a reusable Go module with commonly used tools.

The included tools are:

- read JSON
- write JSON
- produce a JSON encoded error response
- download a static file
- get a random string of length n
- post JSON to a remote service 
- create a directory if it does not already exist

**Not for production -- used in a course.**

## Installation

`go get -u github.com/tsawler/toolbox`

## Usage

```go
package main

import (
	"fmt"
	"github.com/tsawler/toolbox"
)

func main() {
	// create a variable of type toolbox.Tools, so we can use this variable
	// to call the methods on that type
	var tools toolbox.Tools

	// get a random string
	rnd := tools.RandomString(10)
	fmt.Println(rnd)
}
```

### Working with JSON

In a handler, for example:

```go
// JSONPayload is the type for JSON data that we receive
type JSONPayload struct {
    Name string `json:"name"`
    Data string `json:"data"`
}

// jsonResponse is the type for JSON data that we send
type jsonResponse struct {
    Error   bool   `json:"error"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}

// SomeHandler is the handler to accept a post request consisting of json payload
func (app *Config) SomeHandler(w http.ResponseWriter, r *http.Request) {
    var tools toolbox.Tools
    
    // read json into var
    var requestPayload JSONPayload
    _ = tools.ReadJSON(w, r, &requestPayload)
    
    // create the response we'll send back as JSON
    resp := jsonResponse{
        Error:   false,
        Message: "logged",
    }
    
    // write the response back as JSON
    _ = tools.WriteJSON(w, http.StatusAccepted, resp)
}
```

### Download a file

To download a static file, and force it to download instead of displaying
in a browser:

```go
// DownloadAFile downloads an arbitrary file
func (app *Config) DownloadAFile(w http.ResponseWriter, r *http.Request) {
    var tools Tools

    tools.DownloadStaticFile(w, r, "./data", "file.pdf", "file.pdf")
}
```