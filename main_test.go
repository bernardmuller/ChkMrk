package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestRenderItem(t *testing.T) {
	tests := []struct {
		item     Item
		expected string
	}{
		{Item{Index: 1, Name: "Task 1", Completed: true}, "[x] Task 1\n"},
		{Item{Index: 2, Name: "Task 2", Completed: false}, "[ ] Task 2\n"},
	}

	for _, test := range tests {
		var buffer bytes.Buffer
		RenderItemInBuffer(&buffer, test.item)
		actual := buffer.String()

		if actual != test.expected {
			t.Errorf("RenderItem(%v) = %q; expected %q", test.item, actual, test.expected)
		}
	}
}

func TestRenderList(t *testing.T) {
	tests := []struct {
		items    []Item
		expected string
	}{
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
			},
			"[x] Task 1\n[ ] Task 2\n",
		},
		{
			[]Item{
				{Index: 3, Name: "Task 3", Completed: false},
				{Index: 4, Name: "Task 4", Completed: false},
			},
			"[ ] Task 3\n[ ] Task 4\n",
		},
		{
			[]Item{}, // Test for an empty list
			"",
		},
	}

	for _, test := range tests {
		var buffer bytes.Buffer
		RenderListInBuffer(&buffer, test.items)
		actual := buffer.String()

		if actual != test.expected {
			t.Errorf("RenderItems(%v) = %q; expected %q", test.items, actual, test.expected)
		}
	}
}

func TestAddItem(t *testing.T) {
	tests := []struct {
		list     []Item
		item     Item
		expected []Item
	}{
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
			},
			Item{Index: 3, Name: "Task 3", Completed: false},
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
		},

		{
			[]Item{}, // Test for an empty list
			Item{Index: 1, Name: "Task 1", Completed: false},
			[]Item{
				{Index: 1, Name: "Task 1", Completed: false},
			},
		},
	}

	for _, test := range tests {
		actual := AddItemToList(test.list, test.item)

		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("AddItemToList(%v, %v) = %v; expected %v", test.list, test.item, actual, test.expected)
		}
	}
}

func TestCompleteItem(t *testing.T) {
	tests := []struct {
		items    []Item
		index    int
		expected []Item
	}{
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: false},
				{Index: 2, Name: "Task 2", Completed: false},
			},
			1,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
			},
			2,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: true},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
			3,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: true},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
			1,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
		},
	}

	for _, test := range tests {
		actual := CompleteItem(test.items, test.index)

		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("CompleteItem(%v, %v) = %v; expected %v", test.items, test.index, actual, test.expected)
		}
	}
}

func TestInCompleteItem(t *testing.T) {
	tests := []struct {
		items    []Item
		index    int
		expected []Item
	}{
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: false},
				{Index: 2, Name: "Task 2", Completed: false},
			},
			1,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: false},
				{Index: 2, Name: "Task 2", Completed: false},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
			},
			2,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
			1,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: false},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: true},
				{Index: 3, Name: "Task 3", Completed: false},
			},
			2,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
		},
	}

	for index, test := range tests {
		actual := IncompleteItem(test.items, test.index)

		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("Test number %d -> CompleteItem(%v, %v) = %v; expected %v", index, test.items, test.index, actual, test.expected)
		}
	}
}

func TestRemoveItem(t *testing.T) {
	tests := []struct {
		items    []Item
		index    int
		expected []Item
	}{
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: false},
				{Index: 2, Name: "Task 2", Completed: false},
			},
			1,
			[]Item{
				{Index: 2, Name: "Task 2", Completed: false},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
			},
			2,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
			1,
			[]Item{
				{Index: 2, Name: "Task 2", Completed: false},
				{Index: 3, Name: "Task 3", Completed: false},
			},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 2, Name: "Task 2", Completed: true},
				{Index: 3, Name: "Task 3", Completed: false},
			},
			2,
			[]Item{
				{Index: 1, Name: "Task 1", Completed: true},
				{Index: 3, Name: "Task 3", Completed: false},
			},
		},
	}

	for index, test := range tests {
		actual := RemoveItemFromList(test.items, test.index)

		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("Test number %d -> CompleteItem(%v, %v) = %v; expected %v", index, test.items, test.index, actual, test.expected)
		}
	}
}

func TestFindItemInList(t *testing.T) {
	tests := []struct {
		items    []Item
		index    int
		expected Item
	}{
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: false},
				{Index: 2, Name: "Task 2", Completed: false},
        {Index: 3, Name: "Task 3", Completed: false},
        {Index: 4, Name: "Task 4", Completed: false},
        {Index: 5, Name: "Task 5", Completed: false},
        {Index: 6, Name: "Task 6", Completed: false},
        {Index: 7, Name: "Task 7", Completed: false},
        {Index: 8, Name: "Task 8", Completed: false},
			},
			3,
      Item{Index: 3, Name: "Task 3", Completed: false},
		},
		{
			[]Item{
				{Index: 1, Name: "Task 1", Completed: false},
				{Index: 2, Name: "Task 2", Completed: false},
        {Index: 3, Name: "Task 3", Completed: false},
        {Index: 4, Name: "Task 4", Completed: false},
        {Index: 5, Name: "Task 5", Completed: false},
        {Index: 6, Name: "Task 6", Completed: false},
        {Index: 7, Name: "Task 7", Completed: false},
        {Index: 8, Name: "Task 8", Completed: false},
			},
			7,
      Item{Index:  7, Name: "Task 7", Completed: false},
		},
	}

	for index, test := range tests {
		actual, err := FindItemInList(test.items, test.index)

    if(err != nil) {
      t.Errorf("Test number %d -> FindItemInList(%v, %v) = %v; expected %v", index, test.items, test.index, actual, test.expected)
    }

		if !reflect.DeepEqual(actual, test.expected) {
      t.Errorf("Test number %d -> FindItemInList(%v, %v) = %v; expected %v", index, test.items, test.index, actual, test.expected)
		}
	}
}
