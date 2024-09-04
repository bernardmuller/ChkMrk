package main

import (
	// "bytes"
	"database/sql"
	"errors"
	"log"
	"os"
	"reflect"
	// "flag"
	"fmt"
	"io"

	_ "github.com/mattn/go-sqlite3"

	// "log"
	"ChkMrk/cmd"
	"slices"

	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

type Layout int

const (
	Unknown Layout = iota
	Checklists
	ChecklistDetail
	Templates
)

var (
	addItemFlag    string
	checkItemFlag  int
	removeItemFlag int
	listItemsFlag  bool
)

type model struct {
	db              *sql.DB
	items           []Item
	checklists      []Checklist
	choices         []string
	cursor          int
	selected        map[int]Item
	textInput       textinput.Model
	err             error
	showInput       bool
	layout          Layout
	activeList      int
	activeListTitle string
}

func initialModel(db *sql.DB) model {
	items, _ := getItems(db)

	checklists, _ := getChecklists(db)
	choices := make([]string, len(items))
	for i, list := range checklists {
		choices[i] = list.Title
	}
	selected := make(map[int]Item, len(items))
	// for i, item := range items {
	// 	if item.Completed {
	// 		selected[i] = item
	// 	}
	// }

	ti := textinput.New()
	ti.Placeholder = "Steal the moon"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	var currentLayout Layout = Checklists

	return model{
		db:              db,
		items:           items,
		checklists:      checklists,
		choices:         choices,
		selected:        selected,
		textInput:       ti,
		err:             nil,
		showInput:       false,
		layout:          currentLayout,
		activeList:      -1,
		activeListTitle: "",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func InputActionCallback(m *model, msg tea.Msg, cb interface{}, args ...interface{}) (result []reflect.Value, err error) {
	callbackValue := reflect.ValueOf(cb)

	if callbackValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("callback is not a function")
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	result = callbackValue.Call(in)
	return result, nil
}

func HandleInputAction(m *model, msg tea.Msg, handler interface{}) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			m.showInput = false
			return m, nil

		case tea.KeyEnter:
			InputActionCallback(m, msg, handler, m)

			m.textInput.Placeholder = ""
			m.textInput.SetValue("")

			m.showInput = false
			return m, nil

		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd

}

func AddChecklistHandler(m *model) {
	addChecklist(m.db, m.textInput.Value())
	updatedList, _ := getChecklists(m.db)
	m.checklists = updatedList
	choices := make([]string, len(updatedList))
	for i, item := range updatedList {
		choices[i] = item.Title
	}
	m.choices = choices
}

func AddItemHandler(m *model) {
	addItem(m.db, m.textInput.Value(), false, m.activeList)
	updatedList, _ := getItems(m.db)
	m.items = updatedList
	choices := make([]string, len(updatedList))
	for i, item := range updatedList {
		choices[i] = item.Title
	}
	m.choices = choices
}

func ChecklistDetailAction(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showInput {
		return HandleInputAction(&m, msg, AddItemHandler)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

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

		case "x":
			delete(m.selected, m.cursor)
			deleteItem(m.db, m.items[m.cursor].ID)
			updatedList, _ := getItems(m.db)
			m.items = updatedList

			choices := make([]string, len(updatedList))
			for i, item := range updatedList {
				choices[i] = item.Title
			}
			m.choices = choices
			if m.cursor > 1 {
				m.cursor--
			} else {
				m.cursor = 1
			}

		case "n":
			m.showInput = true

		case "esc":
			if m.showInput {
				m.showInput = false
			}

		case "h":
			m.activeList = -1
			lists, _ := getChecklists(m.db)
			m.checklists = lists
			m.layout = 1
			m.cursor = 0

		}
	}

	return m, nil
}

func ChecklistAction(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showInput {
		return HandleInputAction(&m, msg, AddChecklistHandler)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			m.activeList = m.checklists[m.cursor].ID
			m.activeListTitle = m.checklists[m.cursor].Title
			items, _ := getItemsByChecklistId(m.db, m.activeList)
			m.choices = make([]string, len(items))
			for i := 0; i < len(items); i++ {
				m.choices[i] = items[i].Title
				if items[i].Completed {
					m.selected[i] = items[i]
				}
			}
			m.cursor = 0
			m.layout = 2

		case "ctrl+c", "q":
			return m, tea.Quit

		case "n":
			m.showInput = true

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		}
	}

	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.layout {
	case 1:
		return ChecklistAction(m, msg)
	case 2:
		return ChecklistDetailAction(m, msg)
	}
	return m, nil

}

func (m model) View() string {
	switch m.layout {
	case 1:
		return ChecklistView(m)
	case 2:
		return ChecklistDetailView(m)
	}
	return "Not Found\n"
}

func ChecklistView(m model) string {
	s := "\n  My Checklists\n\n"

	for i, list := range m.checklists {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %d. %s\n", cursor, list.ID, list.Title)
	}

	if m.showInput {
		s += fmt.Sprintf(
			"\nEnter title of new checklist:\n\n%s\n\n%s",
			m.textInput.View(),
			"(esc to quit)",
		) + "\n"
	}

	s += "\nPress q to quit.\n"

	return s
}

func ChecklistDetailView(m model) string {
	s := fmt.Sprintf("\n  %s\n\n", m.activeListTitle)

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	if m.showInput {
		s += fmt.Sprintf(
			"\nEnter title of new item:\n\n%s\n\n%s",
			m.textInput.View(),
			"(esc to quit)",
		) + "\n"
	}

	s += "\nPress q to quit.\n"

	return s
}

type Item struct {
	ID          int
	Index       int
	Completed   bool
	Title       string
	ChecklistID int
}

type Checklist struct {
	ID    int
	Title string
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
		title TEXT NOT NULL
	);`

	_, checklistsErr := db.Exec(checklistsQuery)
	if checklistsErr != nil {
		return checklistsErr
	}

	itemsQuery := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL,
    checklist_id INTEGER,
    FOREIGN KEY (checklist_id) REFERENCES checklists(id)
	);
	`
	_, itemsErr := db.Exec(itemsQuery)

	return itemsErr
}

func addItem(db *sql.DB, title string, completed bool, checklist_id int) error {
	query := `INSERT INTO items (title, completed, checklist_id) VALUES (?, ?, ?);`
	_, err := db.Exec(query, title, completed, checklist_id)
	return err
}

func getChecklists(db *sql.DB) ([]Checklist, error) {
	query := `SELECT id, title FROM checklists`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []Checklist
	for rows.Next() {
		var list Checklist
		err := rows.Scan(&list.ID, &list.Title)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
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

func getItemsByChecklistId(db *sql.DB, checklist_id int) ([]Item, error) {
	query := `SELECT id, title, completed FROM items WHERE checklist_id = ?`
	rows, err := db.Query(query, checklist_id)
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

func addChecklist(db *sql.DB, title string) error {
	query := `INSERT INTO checklists (title) VALUES (?);`
	_, err := db.Exec(query, title)
	return err

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

	// itemsQuery := `
	// ALTER TABLE items ADD COLUMN checklist_id INTEGER NOT NULL;
	// `
	// _, err = db.Exec(itemsQuery)
	// if err != nil {
	// 	log.Fatalf("Fail :%s", err.Error())
	// }

	// Initialize the database schema.
	if err := initializeDB(db); err != nil {
		log.Fatalf("Initialization error: %s", err.Error())
	}

	//
	// // NOTE: This will probably be saved inside a DB or a MD file
	seedList := []Item{
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
		{Index: 13, Completed: true, Title: "create new item from addItemFlag arg, and add to list"},
		{Index: 14, Completed: true, Title: "test create new item from arg, completed should be false, index should be list len plus 1"},
		{Index: 15, Completed: false, Title: "create cli process loop"},
		{Index: 16, Completed: false, Title: "render list in process"},
		{Index: 17, Completed: false, Title: "render action prompt underneath rendered list in process"},
		{Index: 18, Completed: false, Title: "capture stdin in process"},
		{Index: 19, Completed: false, Title: "parse stdin command in process prompt"},
		{Index: 20, Completed: false, Title: "map stdin command in process prompt to correct action"},
		{Index: 22, Completed: false, Title: "add identifier to items"},
		{Index: 23, Completed: true, Title: "add index to render"},
		{Index: 21, Completed: false, Title: "save list to sqlite db"},
		{Index: 24, Completed: false, Title: "save list to binary"},
		{Index: 25, Completed: false, Title: "create checklist table"},
		{Index: 26, Completed: false, Title: "create item table"},
		{Index: 27, Completed: false, Title: "create item table"},
		{Index: 28, Completed: false, Title: "refactor flag parsing"},
	}

	checklists, err := getChecklists(db)
	if len(checklists) == 0 {
		err := addChecklist(db, "My First Checklist")
		if err != nil {
			fmt.Printf("Error adding checklist: %s", err.Error())
		}
	}

	dbItems, err := getItems(db)
	if len(dbItems) == 0 {
		firstList, err := getChecklists(db)
		if err != nil {
			log.Printf("ERR: %s", err.Error())
		}
		firstId := firstList[0].ID
		for i := 0; i < len(seedList); i++ {
			err := addItem(db, seedList[i].Title, seedList[i].Completed, firstId)
			if err != nil {
				log.Printf("Error adding item in seed: %s", err.Error())
			}
		}
	}

	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
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
