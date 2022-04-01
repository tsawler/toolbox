[![Version](https://img.shields.io/badge/goversion-1.18.x-blue.svg)](https://golang.org)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/tsawler/goblender/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/tsawler/toolbox)](https://goreportcard.com/report/github.com/tsawler/toolbox)

# Toolbox

A simple example of how to create a reusable Go module with commonly used tools.

**Not for production -- used in a course.**

## Installation

`go get -u github.com/tsawler/toolbox`

## Usage

~~~go
package main

import (
	"fmt"
	"github.com/tsawler/toolbox"
)

func main() {
	var tools toolbox.Tools

	// get a random string
	rnd := tools.RandomString(10)
	fmt.Println(rnd)
}
~~~
