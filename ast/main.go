package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	rawin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	node := Ast(rawin)
	rawout, err := json.MarshalIndent(node, "", "   ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(rawout))
	return
}
