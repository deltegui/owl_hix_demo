//go:build js && wasm

package chart

import (
	"syscall/js"

	"github.com/deltegui/hx"
	"github.com/deltegui/hx/wasm"
)

// instance is a Chart.js chart. Its value is undefined while there is none.
type instance struct{ chart js.Value }

func (c *instance) create(canvas hx.INode, kind, label string, d Data) {
	labels, values := arrays(d)
	c.chart = js.Global().Get("Chart").New(wasm.Underlying(canvas), map[string]any{
		"type": kind,
		"data": map[string]any{
			"labels":   labels,
			"datasets": []any{map[string]any{"label": label, "data": values, "backgroundColor": "rgba(13,110,253,.6)"}},
		},
		"options": map[string]any{"responsive": true, "maintainAspectRatio": false},
	})
}

func (c *instance) destroy() {
	c.chart.Call("destroy")
	c.chart = js.Undefined()
}

func (c *instance) update(d Data) {
	if c.chart.IsUndefined() || c.chart.IsNull() {
		return
	}
	labels, values := arrays(d)
	chartData := c.chart.Get("data")
	chartData.Set("labels", labels)
	chartData.Get("datasets").Index(0).Set("data", values)
	c.chart.Call("update")
}

func arrays(d Data) (labels, values []any) {
	labels, values = make([]any, len(d.Labels)), make([]any, len(d.Values))
	for i, l := range d.Labels {
		labels[i] = l
	}
	for i, v := range d.Values {
		values[i] = v
	}
	return labels, values
}
