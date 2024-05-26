package utils

import (
	"github.com/rivo/tview"
)

type Column struct {
	Text     string
	Align    int
	MaxWidth int
}

func CreateTable(columns []Column, title string) *tview.Table {
	widget := tview.NewTable()
	widget.SetSelectable(true, false)
	widget.SetTitle(title)
	widget.SetBorder(true)
	widget.SetBorders(true)
	for idx, column := range columns {
		widget.SetCell(
			0, idx,
			tview.NewTableCell(column.Text).
				SetSelectable(false).
				SetMaxWidth(column.MaxWidth).
				SetAlign(column.Align),
		)
	}
	widget.SetEvaluateAllRows(true)
	return widget
}

func CreateCell(text string, colWidth int, align int) *tview.TableCell {
	return &tview.TableCell{
		Text:            text,
		NotSelectable:   false,
		Align:           align,
		Color:           tview.Styles.PrimaryTextColor,
		BackgroundColor: tview.Styles.PrimitiveBackgroundColor,
		MaxWidth:        colWidth,
	}
}

func CreateHeaderCell(text string, colWidth int, align int) *tview.TableCell {
	return &tview.TableCell{
		Text:            text,
		NotSelectable:   true,
		Align:           align,
		Color:           tview.Styles.SecondaryTextColor,
		BackgroundColor: tview.Styles.PrimitiveBackgroundColor,
		MaxWidth:        colWidth,
	}
}
