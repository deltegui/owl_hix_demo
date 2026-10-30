// Package chart draws with Chart.js from Go. The page only has to load Chart.js
// (one <script> tag); there is no JavaScript of our own. The data is a signal,
// so the chart follows it. On a server it is the empty canvas, and the browser
// draws the chart once it hydrates.
package chart

import "github.com/deltegui/hx"

type Data struct {
	Labels []string
	Values []float64
}

// New is a canvas with a chart of that kind ("bar", "line", "doughnut"...).
func New(kind, label string, data hx.Gettable[Data]) hx.INode {
	canvas := hx.Canvas()
	var c instance // the chart, while the canvas is in the page

	// The chart is created when the canvas is in the page, and destroyed when it leaves.
	canvas.OnRendered(func() { c.create(canvas, kind, label, hx.UntrackGet(data)) }).OnRemoved(c.destroy)

	// When the data changes, the chart that exists is updated.
	box := hx.Div().Style("position", "relative").Style("height", "260px").Body(canvas)
	box.Effect(func() { c.update(data.Get()) })
	return box
}
