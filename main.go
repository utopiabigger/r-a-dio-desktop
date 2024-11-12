package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/op/paint"
)

// Colors from r-a-d.io
var (
	background = color.NRGBA{R: 20, G: 20, B: 20, A: 255}     // Dark background
	accent     = color.NRGBA{R: 51, G: 181, B: 229, A: 255}   // Blue accent
	textColor  = color.NRGBA{R: 255, G: 255, B: 255, A: 255}  // White text
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("R/a/dio Desktop"),
		)
		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	th := material.NewTheme()
	// Customize theme colors
	th.Palette.Bg = background
	th.Palette.ContrastBg = accent
	th.Palette.Fg = textColor
	
	var ops op.Ops
	var button widget.Clickable

	for e := range w.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			// Fill background
			paint.Fill(gtx.Ops, background)

			if button.Clicked() {
				log.Println("Hello, R/a/dio!")
			}

			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H1(th, "R/a/dio Desktop")
					title.Color = textColor
					title.Alignment = text.Middle
					title.TextSize = unit.Sp(24)
					return title.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(th, &button, "Play Stream")
					btn.Background = accent
					btn.TextSize = unit.Sp(16)
					return layout.Center.Layout(gtx, btn.Layout)
				}),
			)
			e.Frame(gtx.Ops)
		}
	}
	return nil
}