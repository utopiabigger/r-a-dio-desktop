package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/yourusername/radio-desktop/audio"
)

// Colors from r-a-d.io
var (
	background = color.NRGBA{R: 20, G: 20, B: 20, A: 255}    // Dark background
	accent     = color.NRGBA{R: 51, G: 181, B: 229, A: 255}  // Blue accent
	textColor  = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White text
)

const streamURL = "https://relay0.r-a-d.io/main.mp3"

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
	th.Palette.Bg = background
	th.Palette.ContrastBg = accent
	th.Palette.Fg = textColor

	var ops op.Ops
	var button widget.Clickable

	// Initialize audio player
	audioPlayer, err := audio.NewPlayer()
	if err != nil {
		return err
	}

	for e := range w.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			paint.Fill(gtx.Ops, background)

			if button.Clicked() {
				if !audioPlayer.IsPlaying() {
					if err := audioPlayer.PlayStream(streamURL); err != nil {
						log.Printf("Error starting stream: %v", err)
					}
				} else {
					audioPlayer.Stop()
				}
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
					buttonText := "Play Stream"
					if audioPlayer.IsPlaying() {
						buttonText = "Stop Stream"
					}
					btn := material.Button(th, &button, buttonText)
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
