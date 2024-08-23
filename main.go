package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
)

var (
	addFlag    string
	checkFlag  int
	removeFlag int
	listFlag   bool
)

type Item struct {
	Index     int
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
	list = append(list, item)
	return list
}

func CompleteItem(list []Item, index int) []Item {
	for i := 0; i < len(list); i++ {
		if list[i].Index == index {
			list[i].Completed = true
		}
	}
	return list
}

func IncompleteItem(list []Item, index int) []Item {
	for i := 0; i < len(list); i++ {
		if list[i].Index == index {
			list[i].Completed = false
		}
	}
	return list
}

func RemoveItemFromList(list []Item, index int) []Item {
	var itemSliceIndex int
	for i := 0; i < len(list); i++ {
		if list[i].Index == index {
			itemSliceIndex = i
		}
	}
	list = slices.Delete(list, itemSliceIndex, itemSliceIndex+1)
	return list
}

func FindItemInList(list []Item, index int) (Item, error) {
	for i := 0; i < len(list); i++ {
		if list[i].Index == index {
			return list[i], nil
		}
	}
	return Item{}, errors.New(fmt.Sprintf("Failed to find Item with Index: %d", index))
}

func main() {

	// NOTE: This will probably be saved inside a DB or a MD file
	list := []Item{
		{Index: 1, Completed: true, Name: "make a list of items"},
		{Index: 2, Completed: true, Name: "test functions that render items"},
		{Index: 3, Completed: true, Name: "parse cli entrypoint without args"},
		{Index: 4, Completed: true, Name: "parse cli entrypoint with args"},
		{Index: 5, Completed: true, Name: "parse add item flag"},
		{Index: 12, Completed: true, Name: "parse check item flag"},
		{Index: 11, Completed: true, Name: "parse remove item flag"},
		{Index: 10, Completed: true, Name: "list args"},
		{Index: 6, Completed: true, Name: "test check an item"},
		{Index: 7, Completed: true, Name: "test uncheck an item"},
		{Index: 8, Completed: true, Name: "test add a new item"},
		{Index: 9, Completed: true, Name: "test remove an item"},
		{Index: 13, Completed: false, Name: "create new item from addFlag arg, and add to list"},
		{Index: 14, Completed: false, Name: "test create new item from arg, completed should be false, index should be list len plus 1"},
		{Index: 15, Completed: false, Name: "create cli process loop"},
		{Index: 16, Completed: false, Name: "render list in process"},
		{Index: 17, Completed: false, Name: "render action prompt underneath rendered list in process"},
		{Index: 18, Completed: false, Name: "capture stdin in process"},
		{Index: 19, Completed: false, Name: "parse stdin command in process prompt"},
		{Index: 20, Completed: false, Name: "map stdin command in process prompt to correct action"},
		{Index: 21, Completed: false, Name: "save list to .txt file"},
		{Index: 22, Completed: false, Name: "add identifier to items"},
		{Index: 23, Completed: false, Name: "add index to render"},
	}

	// NOTE: We can probably extract this to a new function
	flag.StringVar(&addFlag, "a", "default", "help message")
	flag.IntVar(&checkFlag, "c", 0, "help message")
	flag.IntVar(&removeFlag, "r", 0, "help message")
	flag.BoolVar(&listFlag, "l", false, "help message")
	flag.Parse()

	var buffer bytes.Buffer

	if addFlag != "default" {
		newlist := AddItemToList(list, Item{Index: len(list) + 1, Completed: false, Name: addFlag})

		RenderListInBuffer(&buffer, newlist)
		fmt.Print(buffer.String())
	}

	if checkFlag != 0 {
		item, err := FindItemInList(list, checkFlag)

		if err != nil {
			fmt.Errorf("Error checking item: %s", err.Error())
		}

		if !item.Completed {
			list = CompleteItem(list, item.Index)
			RenderListInBuffer(&buffer, list)
			fmt.Print(buffer.String())
		} else {
			list = IncompleteItem(list, item.Index)
			RenderListInBuffer(&buffer, list)
			fmt.Print(buffer.String())
		}
	}

	if removeFlag != 0 {
		item, err := FindItemInList(list, removeFlag)

		if err != nil {
			fmt.Errorf("Error checking item: %s", err.Error())
		}

		list = RemoveItemFromList(list, item.Index)
		RenderListInBuffer(&buffer, list)
		fmt.Print(buffer.String())
	}

	if listFlag {
		RenderListInBuffer(&buffer, list)
		fmt.Print(buffer.String())
	}
}
