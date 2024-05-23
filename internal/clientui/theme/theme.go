package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var BaseTheme tview.Theme = tview.Theme{
	PrimitiveBackgroundColor:    tcell.ColorBlack,
	ContrastBackgroundColor:     tcell.ColorBlue,
	MoreContrastBackgroundColor: tcell.ColorGreen,
	BorderColor:                 tcell.ColorWhite,
	TitleColor:                  tcell.ColorWhite,
	GraphicsColor:               tcell.ColorWhite,
	PrimaryTextColor:            tcell.ColorWhite,
	SecondaryTextColor:          tcell.ColorYellow,
	TertiaryTextColor:           tcell.ColorGreen,
	InverseTextColor:            tcell.ColorBlue,
	ContrastSecondaryTextColor:  tcell.ColorNavy,
}

const (
	LabelColor           = tcell.ColorGreen
	FocusedBorderColor   = tcell.ColorDarkCyan
	UnfocusedBorderColor = tcell.ColorWhite
	TabColor             = tcell.ColorBlack
	TabBackgroundColor   = tcell.ColorGray
)

func ApplyTheme() {
	tview.Styles = BaseTheme

	tview.Borders.BottomLeft = tview.BoxDrawingsLightArcUpAndRight
	tview.Borders.BottomRight = tview.BoxDrawingsLightArcUpAndLeft
	tview.Borders.TopLeft = tview.BoxDrawingsLightArcDownAndRight
	tview.Borders.TopRight = tview.BoxDrawingsLightArcDownAndLeft

	tview.Borders.VerticalFocus = tview.Borders.Vertical
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal

	tview.Borders.BottomLeftFocus = tview.Borders.BottomLeft
	tview.Borders.BottomRightFocus = tview.Borders.BottomRight
	tview.Borders.TopLeftFocus = tview.Borders.TopLeft
	tview.Borders.TopRightFocus = tview.Borders.TopRight
}
