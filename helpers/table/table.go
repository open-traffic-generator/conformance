package table

import (
	"fmt"
	"strings"
)

type Table struct {
	Title    string
	Headers  []string
	Rows     [][]interface{}
	ColWidth int
}

func NewTable(title string, headers []string, colWidth int) *Table {
	return &Table{
		Title:    title,
		Headers:  headers,
		Rows:     make([][]interface{}, 0, 4),
		ColWidth: colWidth,
	}
}

func (tb *Table) AppendRow(row []interface{}) {
	diff := len(tb.Headers) - len(row)
	for i := 0; i < diff; i++ {
		row = append(row, "_")
	}

	tb.Rows = append(tb.Rows, row)
}

func (tb *Table) String() string {
	var out strings.Builder

	border := strings.Repeat("-", tb.ColWidth*len(tb.Headers))
	out.WriteString("\n")
	out.WriteString(border)
	out.WriteString(fmt.Sprintf("\n%s\n", tb.Title))
	out.WriteString(border)
	out.WriteString("\n")
	for _, h := range tb.Headers {
		out.WriteString(fmt.Sprintf("%-15s", h))
	}
	out.WriteString("\n")

	for _, row := range tb.Rows {
		for _, col := range row {
			out.WriteString(fmt.Sprintf("%-15v", col))
		}
		out.WriteString("\n")
	}
	out.WriteString(border)
	out.WriteString("\n\n")

	return out.String()
}
