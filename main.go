package main

import (
	// "bytes"
	"database/sql"
	"errors"
	"log"
	"os"

	// "flag"
	"fmt"
	"io"

	_ "github.com/mattn/go-sqlite3"

	// "log"
	"ChkMrk/cmd"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	addItemFlag    string
	checkItemFlag  int
	removeItemFlag int
	listItemsFlag  bool
)

type model struct {
	db       *sql.DB
	items    []Item
	choices  []string     // items on the to-do list
	cursor   int          // which to-do list item our cursor is pointing at
	selected map[int]Item // which to-do items are selected
}

func initialModel(db *sql.DB) model {
	items, _ := getItems(db)
	choices := make([]string, len(items))
	for i, item := range items {
		choices[i] = item.Title
	}
	selected := make(map[int]Item, len(items))
	for i, item := range items {
		if item.Completed {
			selected[i] = item
		}
	}

	return model{
		db:       db,
		items:    items,
		choices:  choices,
		selected: selected,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
				updateItemCompleted(m.db, m.items[m.cursor].ID, false)
				updatedList, _ := getItems(m.db)
				m.items = updatedList
			} else {
				m.selected[m.cursor] = m.items[m.cursor]
				updateItemCompleted(m.db, m.items[m.cursor].ID, true)
				updatedList, _ := getItems(m.db)
				m.items = updatedList
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	s := "My Checklist\n\n"

	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}

type Item struct {
	ID        int
	Index     int
	Completed bool
	Title     string
}

var list []Item

func RenderItemInBuffer(w io.Writer, item Item) {
	if item.ID > 9 {
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
	checklistsQuery := `
	CREATE TABLE IF NOT EXISTS checklists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL
	);`

	_, checklistsErr := db.Exec(checklistsQuery)
	if checklistsErr != nil {
		return checklistsErr
	}

	itemsQuery := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL
	);
	`
	_, itemsErr := db.Exec(itemsQuery)
	return itemsErr
}

func addItem(db *sql.DB, title string, completed bool) error {
	query := `INSERT INTO items (title, completed) VALUES (?, ?)`
	_, err := db.Exec(query, title, completed)
	return err
}

func getItems(db *sql.DB) ([]Item, error) {
	query := `SELECT id, title, completed FROM items`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Title, &item.Completed)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func updateItemCompleted(db *sql.DB, id int, completed bool) error {
	query := `UPDATE items SET completed = ? WHERE id = ?`
	_, err := db.Exec(query, completed, id)
	return err
}

func getItemById(db *sql.DB, id int) (Item, error) {
	query := `SELECT * FROM items WHERE id = ?`

	row, err := db.Query(query, id)
	if err != nil {
		return Item{}, err
	}
	defer row.Close()

	var item Item

	for row.Next() {
		err = row.Scan(&item.ID, &item.Title, &item.Completed)
		if err != nil {
			return Item{}, err
		}
	}

	return item, nil
}

func deleteItem(db *sql.DB, id int) error {
	query := `DELETE FROM items WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func main() {

	cmd.Execute()

	db, err := sql.Open("sqlite3", "./checklist.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize the database schema.
	if err := initializeDB(db); err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	//
	// // NOTE: This will probably be saved inside a DB or a MD file
	// seedList := []Item{
	// 	{Index: 1, Completed: true, Title: "make a list of items"},
	// 	{Index: 2, Completed: true, Title: "test functions that render items"},
	// 	{Index: 3, Completed: true, Title: "parse cli entrypoint without args"},
	// 	{Index: 4, Completed: true, Title: "parse cli entrypoint with args"},
	// 	{Index: 5, Completed: true, Title: "parse add item flag"},
	// 	{Index: 12, Completed: true, Title: "parse check item flag"},
	// 	{Index: 11, Completed: true, Title: "parse remove item flag"},
	// 	{Index: 10, Completed: true, Title: "list args"},
	// 	{Index: 6, Completed: true, Title: "test check an item"},
	// 	{Index: 7, Completed: true, Title: "test uncheck an item"},
	// 	{Index: 8, Completed: true, Title: "test add a new item"},
	// 	{Index: 9, Completed: true, Title: "test remove an item"},
	// 	{Index: 13, Completed: true, Title: "create new item from addItemFlag arg, and add to list"},
	// 	{Index: 14, Completed: true, Title: "test create new item from arg, completed should be false, index should be list len plus 1"},
	// 	{Index: 15, Completed: false, Title: "create cli process loop"},
	// 	{Index: 16, Completed: false, Title: "render list in process"},
	// 	{Index: 17, Completed: false, Title: "render action prompt underneath rendered list in process"},
	// 	{Index: 18, Completed: false, Title: "capture stdin in process"},
	// 	{Index: 19, Completed: false, Title: "parse stdin command in process prompt"},
	// 	{Index: 20, Completed: false, Title: "map stdin command in process prompt to correct action"},
	// 	{Index: 22, Completed: false, Title: "add identifier to items"},
	// 	{Index: 23, Completed: true, Title: "add index to render"},
	// 	{Index: 21, Completed: false, Title: "save list to sqlite db"},
	// 	{Index: 24, Completed: false, Title: "save list to binary"},
	// 	{Index: 25, Completed: false, Title: "create checklist table"},
	// 	{Index: 26, Completed: false, Title: "create item table"},
	// 	{Index: 27, Completed: false, Title: "create item table"},
	// 	{Index: 28, Completed: false, Title: "refactor flag parsing"},
	// }
	//
	// dbItems, err := getItems(db)
	// if len(dbItems) == 0 {
	// 	for i := 0; i < len(seedList); i++ {
	// 		addItem(db, list[i].Title, list[i].Completed)
	// 	}
	// }
	//
	// // NOTE: We can probably extract this to a new function
	// flag.StringVar(&addItemFlag, "a", "default", "help message")
	// flag.IntVar(&checkItemFlag, "c", 0, "help message")
	// flag.IntVar(&removeItemFlag, "r", 0, "help message")
	// flag.BoolVar(&listItemsFlag, "l", false, "help message")
	// flag.Parse()
	//
	// var buffer bytes.Buffer
	//
	// if addItemFlag != "default" {
	// 	err := addItem(db, addItemFlag, false)
	// 	if err != nil {
	// 		log.Fatalf("Error adding item: %s", err.Error())
	// 	}
	//
	// 	list, err := getItems(db)
	// 	if err != nil {
	// 		log.Printf("Error getting items after adding new item")
	// 	}
	//
	// 	RenderListInBuffer(&buffer, list)
	// 	fmt.Print(buffer.String())
	// }
	//
	// if checkItemFlag != 0 {
	// 	item, err := getItemById(db, checkItemFlag)
	// 	if err != nil {
	// 		log.Fatalf("Error finding item with id %d: %s", checkItemFlag, err.Error())
	// 	}
	//
	// 	if !item.Completed {
	// 		updateItemCompleted(db, item.ID, true)
	// 		list, err := getItems(db)
	// 		if err != nil {
	// 			fmt.Errorf("Error updating item: %s", err.Error())
	// 		}
	// 		RenderListInBuffer(&buffer, list)
	// 		fmt.Print(buffer.String())
	// 	} else {
	// 		updateItemCompleted(db, item.ID, false)
	// 		list, err := getItems(db)
	// 		if err != nil {
	// 			fmt.Errorf("Error updating item: %s", err.Error())
	// 		}
	// 		RenderListInBuffer(&buffer, list)
	// 		fmt.Print(buffer.String())
	// 	}
	// }
	//
	// if removeItemFlag != 0 {
	// 	item, err := getItemById(db, removeItemFlag)
	// 	if err != nil {
	// 		log.Fatalf("Error finding item with ID %d: %s", removeItemFlag, err.Error())
	// 	}
	//
	// 	list, err := getItems(db)
	// 	if err != nil {
	// 		fmt.Errorf("Error updating item: %s", err.Error())
	// 	}
	//
	// 	err = deleteItem(db, item.ID)
	// 	if err != nil {
	// 		log.Fatalf("Error deleteing item: %s", err.Error())
	// 	}
	//
	// 	list, err = getItems(db)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	//
	// 	RenderListInBuffer(&buffer, list)
	// 	fmt.Print(buffer.String())
	// }
	//
	// if listItemsFlag {
	// 	dbList, err := getItems(db)
	// 	if err != nil {
	// 		log.Fatalf("Unable to list items: %s", err)
	// 	}
	// 	RenderListInBuffer(&buffer, dbList)
	// 	fmt.Print(buffer.String())
	// }
}
