package main

import (
	"bytes"
	"testing"
)

func TestRenderItem(t *testing.T) {
	tests := []struct {
		item     Item
		expected string
	}{
		{Item{Name: "Task 1", Completed: true}, "[x] Task 1\n"},
		{Item{Name: "Task 2", Completed: false}, "[ ] Task 2\n"},
	}

	for _, test := range tests {
		var buffer bytes.Buffer
		RenderItem(&buffer, test.item)
		actual := buffer.String()

		if actual != test.expected {
			t.Errorf("RenderItem(%v) = %q; expected %q", test.item, actual, test.expected)
		}
	}
}
