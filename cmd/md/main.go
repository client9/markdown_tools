package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/client9/markdown_tools"

	gfm "github.com/shurcooL/github_flavored_markdown"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	vetCommand    = kingpin.Command("vet", "vet markdown structure")
	fmtComment    = kingpin.Command("fmt", "reformat markdown")
	renderCommand = kingpin.Command("render", "render markdown to another format")
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
		rawout := gfm.Markdown(rawin)
		os.Stdout.Write(rawout)
	}
}
