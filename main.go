package main

import (
	"bytes"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"slices"
)

var (
	addFlag    string
	checkFlag  int
	removeFlag int
	listFlag   bool
)

type Item struct {
	ID        int
	Index     int
	Completed bool
	Title     string
}

var list []Item

func RenderItemInBuffer(w io.Writer, item Item) {
	if item.Index > 9 {
		fmt.Fprintf(w, "%d. ", item.ID)
	} else {
		fmt.Fprintf(w, "%d.  ", item.ID)
	}
	if item.Completed {
		fmt.Fprint(w, "[x]")
	} else {
		fmt.Fprint(w, "[ ]")
	}
	fmt.Fprintf(w, " %s\n", item.Title)
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

func initializeDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

func addItem(db *sql.DB, title string, completed bool) error {
	query := `INSERT INTO items (title, completed) VALUES (?, ?)`
	_, err := db.Exec(query, title, completed)
	return err
}

func main() {

	db, err := sql.Open("sqlite3", "./checklist.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize the database schema.
	if err := initializeDB(db); err != nil {
		log.Fatal(err)
	}

	// NOTE: This will probably be saved inside a DB or a MD file
	list := []Item{
		{Index: 1, Completed: true, Title: "make a list of items"},
		{Index: 2, Completed: true, Title: "test functions that render items"},
		{Index: 3, Completed: true, Title: "parse cli entrypoint without args"},
		{Index: 4, Completed: true, Title: "parse cli entrypoint with args"},
		{Index: 5, Completed: true, Title: "parse add item flag"},
		{Index: 12, Completed: true, Title: "parse check item flag"},
		{Index: 11, Completed: true, Title: "parse remove item flag"},
		{Index: 10, Completed: true, Title: "list args"},
		{Index: 6, Completed: true, Title: "test check an item"},
		{Index: 7, Completed: true, Title: "test uncheck an item"},
		{Index: 8, Completed: true, Title: "test add a new item"},
		{Index: 9, Completed: true, Title: "test remove an item"},
		{Index: 13, Completed: true, Title: "create new item from addFlag arg, and add to list"},
		{Index: 14, Completed: true, Title: "test create new item from arg, completed should be false, index should be list len plus 1"},
		//
		// cli process
		//
		{Index: 15, Completed: false, Title: "create cli process loop"},
		{Index: 16, Completed: false, Title: "render list in process"},
		{Index: 17, Completed: false, Title: "render action prompt underneath rendered list in process"},
		{Index: 18, Completed: false, Title: "capture stdin in process"},
		{Index: 19, Completed: false, Title: "parse stdin command in process prompt"},
		{Index: 20, Completed: false, Title: "map stdin command in process prompt to correct action"},
		{Index: 22, Completed: false, Title: "add identifier to items"},
		//
		{Index: 23, Completed: true, Title: "add index to render"},
		//
		// persistence
		//
		{Index: 21, Completed: false, Title: "save list to sqlite db"},
		{Index: 24, Completed: false, Title: "save list to binary"},
		{Index: 25, Completed: false, Title: "create checklist table"},
		{Index: 26, Completed: false, Title: "create item table"},
		{Index: 27, Completed: false, Title: "create item table"},
		//
		// config
		//
		// yaml ?
		// json ?
		// txt ?
		{Index: 28, Completed: false, Title: "refactor flag parsing"},
	}

	// NOTE: We can probably extract this to a new function
	flag.StringVar(&addFlag, "a", "default", "help message")
	flag.IntVar(&checkFlag, "c", 0, "help message")
	flag.IntVar(&removeFlag, "r", 0, "help message")
	flag.BoolVar(&listFlag, "l", false, "help message")
	flag.Parse()

	var buffer bytes.Buffer

	if addFlag != "default" {
		newlist := AddItemToList(list, Item{Index: len(list) + 1, Completed: false, Title: addFlag})

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
