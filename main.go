package main

import "github.com/mdb/gh-dispatch/cmd"

// version's value is passed in at build time.
var version string

func main() {
	cmd.Execute(version)
}
