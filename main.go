package main

import "fmt"

type Item struct {
	Completed bool
	Name      string
}

var list []Item

func RenderList(list []Item) {
	for i := 0; i < len(list); i++ {
		fmt.Println(list[i].Name)
	}
}

func main() {
	list = append(list, Item{Completed: false, Name: "make a list of items"})
	list = append(list, Item{Completed: false, Name: "check an item"})
	list = append(list, Item{Completed: false, Name: "uncheck an item"})
	list = append(list, Item{Completed: false, Name: "add a new item"})
	list = append(list, Item{Completed: false, Name: "remove an item"})
	RenderList(list)
}
