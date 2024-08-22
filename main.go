package main

import (
	"bytes"
	"fmt"
	"io"
)

type Item struct {
	Completed bool
	Name      string
}

var list []Item

func RenderItem(w io.Writer, item Item) {
	if item.Completed {
		fmt.Fprint(w, "[x]")
	} else {
		fmt.Fprint(w, "[ ]")
	}
	fmt.Fprintf(w, " %s\n", item.Name)
}

func RenderList(w io.Writer, list []Item) {
	for i := 0; i < len(list); i++ {
		RenderItem(w, list[i])
	}
}

func main() {
	list = append(list, Item{Completed: true, Name: "make a list of items"})
	list = append(list, Item{Completed: true, Name: "test functions that render items"})
	list = append(list, Item{Completed: false, Name: "parse cli entrypoint without args"})
	list = append(list, Item{Completed: false, Name: "parse cli entrypoint with args"})
	list = append(list, Item{Completed: false, Name: "list args"})
	list = append(list, Item{Completed: false, Name: "test check an item"})
	list = append(list, Item{Completed: false, Name: "test uncheck an item"})
	list = append(list, Item{Completed: false, Name: "test add a new item"})
	list = append(list, Item{Completed: false, Name: "test remove an item"})
	var buffer bytes.Buffer
	RenderList(&buffer, list)
}
