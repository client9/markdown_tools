package main

import (
	//	"github.com/client9/markdown_tools"
	"io/ioutil"
	"log"
	"os"

	//"go4.org/errorutil"
	gfm "github.com/shurcooL/github_flavored_markdown"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	renderCommand = kingpin.Command("render", "render markdown to another format")
)

func main() {
	switch kingpin.Parse() {
	case "fmt":

	case "vet:

	case "render":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		rawout := gfm.Markdown(rawin)
		os.Stdout.Write(rawout)
	}
}
