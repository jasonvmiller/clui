package clui

import (
	"fmt"
	// xs "github.com/huandu/xstrings"
	term "github.com/nsf/termbox-go"
)

/*
BarChart is a chart that represents grouped data with
rectangular bars. It can be monochrome - defaut behavior.
One can assign individual color to each bar and even use
custom drawn bars to display multicolored bars depending
on bar value.
All bars have the same width: either constant BarSize - in
case of AutoSize is false, or automatically calculated but
cannot be less than BarSize. Bars that do not fit the chart
area are not displayed.
BarChart displays vertical axis with values on the chart left
if ValueWidth greater than 0, horizontal axis with bar titles
if ShowTitles is true (to enable displaying marks on horizontal
axis, set ShowMarks to true), and chart legend on the right if
LegendWidth is greater than 3.
If LegendWidth is greater than half of the chart it is not
displayed. The same is applied to ValueWidth
*/
type SparkChart struct {
	ControlBase
	data         []float64
	valueWidth   int
	hiliteMax    bool
	maxFg, maxBg term.Attribute
	topValue     float64
	autosize     bool
}

/*
NewSparkChart creates a new spark chart.
view - is a View that manages the control
parent - is container that keeps the control. The same View can be a view and a parent at the same time.
w and h - are minimal size of the control.
scale - the way of scaling the control when the parent is resized. Use DoNotScale constant if the
control should keep its original size.
*/
func NewSparkChart(view View, parent Control, w, h int, scale int) *SparkChart {
	c := new(SparkChart)

	if w == AutoSize {
		w = 10
	}
	if h == AutoSize {
		h = 5
	}

	c.view = view
	c.parent = parent

	c.SetSize(w, h)
	c.SetConstraints(w, h)
	c.tabSkip = true
	c.hiliteMax = true
	c.autosize = true
	c.data = make([]float64, 0)

	if parent != nil {
		parent.AddChild(c, scale)
	}

	return c
}

// Repaint draws the control on its View surface
func (b *SparkChart) Repaint() {
	canvas := b.view.Canvas()
	tm := b.view.Screen().Theme()

	fg, bg := RealColor(tm, b.fg, ColorSparkChartText), RealColor(tm, b.bg, ColorSparkChartBack)
	canvas.FillRect(b.x, b.y, b.width, b.height, term.Cell{Ch: ' ', Fg: fg, Bg: bg})

	if len(b.data) == 0 {
		return
	}

	b.drawValues(fg, bg)
	b.drawBars(tm)
}

func (b *SparkChart) drawBars(tm Theme) {
	if len(b.data) == 0 {
		return
	}

	start, width := b.calculateBarArea()
	if width < 2 {
		return
	}

	coeff, max := b.calculateMultiplier()
	if coeff == 0.0 {
		return
	}

	h := b.height
	pos := b.x + start
	canvas := b.view.Canvas()

	mxFg, mxBg := RealColor(tm, b.maxFg, ColorSparkChartMaxText), RealColor(tm, b.maxBg, ColorSparkChartMaxBack)
	brFg, brBg := RealColor(tm, b.fg, ColorSparkChartBarText), RealColor(tm, b.bg, ColorSparkChartBarBack)
	parts := []rune(tm.SysObject(ObjSparkChart))

	var dt []float64
	if len(b.data) > width {
		dt = b.data[len(b.data)-width:]
	} else {
		dt = b.data
	}

	for _, d := range dt {
		barH := int(d * coeff)

		if barH <= 0 {
			pos++
			continue
		}

		f, g := brFg, brBg
		if b.hiliteMax && max == d {
			f, g = mxFg, mxBg
		}
		cell := term.Cell{Ch: parts[0], Fg: f, Bg: g}
		canvas.FillRect(pos, b.y+h-barH, 1, barH, cell)

		pos++
	}
}

func (b *SparkChart) drawValues(fg, bg term.Attribute) {
	if b.valueWidth <= 0 {
		return
	}

	pos, _ := b.calculateBarArea()
	if pos == 0 {
		return
	}

	h := b.height
	coeff, max := b.calculateMultiplier()
	if max == coeff {
		return
	}
	if !b.autosize || b.topValue == 0 {
		max = b.topValue
	}

	canvas := b.view.Canvas()
	dy := 0
	format := fmt.Sprintf("%%%v.2f", b.valueWidth)
	for dy < h-1 {
		v := float64(h-dy) / float64(h) * max
		s := fmt.Sprintf(format, v)
		s = CutText(s, b.valueWidth)
		canvas.PutText(b.x, b.y+dy, s, fg, bg)

		dy += 2
	}
}

func (b *SparkChart) calculateBarArea() (int, int) {
	w := b.width
	pos := 0

	if b.valueWidth < w/2 {
		w = w - b.valueWidth
		pos = b.valueWidth
	}

	return pos, w
}

func (b *SparkChart) calculateMultiplier() (float64, float64) {
	if len(b.data) == 0 {
		return 0, 0
	}

	h := b.height
	if h <= 1 {
		return 0, 0
	}

	max := b.data[0]
	for _, val := range b.data {
		if val > max {
			max = val
		}
	}

	if max == 0 {
		return 0, 0
	}

	if b.autosize || b.topValue == 0 {
		return float64(h) / max, max
	} else {
		return float64(h) / b.topValue, max
	}
}

// AddData appends a new bar to a chart
func (b *SparkChart) AddData(val float64) {
	b.data = append(b.data, val)

	_, width := b.calculateBarArea()
	if len(b.data) > width {
		b.data = b.data[len(b.data)-width:]
	}
	b.Logger().Printf("%v - %v = %v", b.width, width, len(b.data))
}

// ClearData removes all bar from chart
func (b *SparkChart) ClearData() {
	b.data = make([]float64, 0)
}

// SetData assign a new bar list to a chart
func (b *SparkChart) SetData(data []float64) {
	b.data = make([]float64, len(data))
	copy(b.data, data)

	_, width := b.calculateBarArea()
	if len(b.data) > width {
		b.data = b.data[len(b.data)-width:]
	}
}

// ValueWidth returns the width of the area at the left of
// chart used to draw values. Set it to 0 to turn off the
// value panel
func (b *SparkChart) ValueWidth() int {
	return b.valueWidth
}

// SetValueWidth changes width of the value panel on the left
func (b *SparkChart) SetValueWidth(width int) {
	b.valueWidth = width
}

// Top returns the value of the top of a chart. The value is
// used only if autosize is off to scale all the data
func (b *SparkChart) Top() float64 {
	return b.topValue
}

// SetTop sets the theoretical highest value of data flow
// to scale the chart
func (b *SparkChart) SetTop(top float64) {
	b.topValue = top
}

// AutoScale returns whether spark chart scales automatically
// depending on displayed data or it scales using Top value
func (b *SparkChart) AutoScale() bool {
	return b.autosize
}

// SetAutoScale changes the way of scaling the data flow
func (b *SparkChart) SetAutoScale(auto bool) {
	b.autosize = auto
}

// HilitePeaks returns whether chart draws maximum peaks
// with different color
func (b *SparkChart) HilitePeaks() bool {
	return b.hiliteMax
}

// SetHilitePeaks enables or disables hiliting maximum
// values with different colors
func (b *SparkChart) SetHilitePeaks(hilite bool) {
	b.hiliteMax = hilite
}