# Toolbox

A simple example of how to create a reusable Go module with commonly used tools.

**Not for production -- used in a course.**

## Installation

`go get -u github.com/tsawler/toolbox`

## Usage

~~~
package main

import (
	"fmt"
	"github.com/tsawler/toolbox"
)

func main() {
	var tools toolbox.Tools

	rnd := tools.RandomString(10)
	fmt.Println(rnd)
}
~~~
