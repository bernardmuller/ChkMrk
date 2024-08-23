package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
)

var (
	addFlag string
)

type Item struct {
	Completed bool
	Name      string
}

var list []Item

func RenderItemInBuffer(w io.Writer, item Item) {
	if item.Completed {
		fmt.Fprint(w, "[x]")
	} else {
		fmt.Fprint(w, "[ ]")
	}
	fmt.Fprintf(w, " %s\n", item.Name)
}

func RenderListInBuffer(w io.Writer, list []Item) {
	for i := 0; i < len(list); i++ {
		RenderItemInBuffer(w, list[i])
	}
}

func AddItemToList(list []Item, item Item) []Item {
	list = append(list, Item{Completed: false, Name: addFlag})
	list = append(list, item)
	return list
}

func main() {

	//NOTE: This will probably be saved inside a DB or a MD file
	// NOTE: This will probably be saved inside a DB or a MD file
	list := []Item{
		{Completed: true, Name: "make a list of items"},
		{Completed: true, Name: "test functions that render items"},
		{Completed: true, Name: "parse cli entrypoint without args"},
		{Completed: true, Name: "parse cli entrypoint with args"},
		{Completed: true, Name: "parse add item flag"},
		{Completed: false, Name: "parse check item flag"},
		{Completed: false, Name: "parse remove item flag"},
		{Completed: false, Name: "list args"},
		{Completed: false, Name: "test check an item"},
		{Completed: false, Name: "test uncheck an item"},
		{Completed: false, Name: "test add a new item"},
		{Completed: false, Name: "test remove an item"},
	}

	// NOTE: We can probably extract this to a new function
	flag.StringVar(&addFlag, "a", "default", "help message")
	flag.Parse()

	var buffer bytes.Buffer

	if addFlag == "default" {
		RenderListInBuffer(&buffer, list)
		fmt.Print(buffer.String())
	} else {
		newlist := AddItemToList(list, Item{Completed: false, Name: addFlag})

		RenderListInBuffer(&buffer, newlist)
		fmt.Print(buffer.String())
	}
}
