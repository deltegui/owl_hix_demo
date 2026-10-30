//go:build !(js && wasm)

package chart

import "github.com/deltegui/hx"

// instance does nothing on a server: Chart.js only exists in the browser.
type instance struct{}

func (*instance) create(hx.INode, string, string, Data) {}
func (*instance) destroy()                              {}
func (*instance) update(Data)                           {}
