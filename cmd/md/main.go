package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/client9/markdown_tools"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	vetCommand    = kingpin.Command("vet", "vet markdown structure")
	fmtComment    = kingpin.Command("fmt", "reformat markdown")
	renderCommand = kingpin.Command("render", "render markdown to another format")
	renderType    = renderCommand.Arg("type", "render type").Default("html").String()
)

func main() {
	switch kingpin.Parse() {
	case "fmt":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		out := mdtool.Fmt(rawin, nil)
		fmt.Println(string(out))
	case "vet":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		faults := mdtool.Vet(rawin)
		for _, f := range faults {
			fmt.Printf("%d:%d offset=%d reason=%s %q\n", f.Row, f.Column, f.Offset, f.Reason, f.Line)
		}
		if len(faults) > 0 {
			os.Exit(2)
		}
	case "render":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		var rawout []byte
		switch *renderType {
		case "html":
			rawout = mdtool.RenderHTML(rawin)
		case "github":
			rawout = mdtool.RenderGitHub(rawin)
		default:
			log.Fatalf("Unknown render type %q", *renderType)
		}
		os.Stdout.Write(rawout)
	}
}
