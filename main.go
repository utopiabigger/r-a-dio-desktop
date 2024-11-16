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
	"image/png"
)

// Colors from r-a-d.io
var (
	background = color.NRGBA{R: 34, G: 34, B: 34, A: 255}    // Updated dark background
	accent     = color.NRGBA{R: 51, G: 181, B: 229, A: 255}  // Blue accent
	textColor  = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White text
)

const streamURL = "https://relay0.r-a-d.io/main.mp3"

var (
	volumeSlider widget.Float
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("r/a/dio Desktop"),
			app.Size(unit.Dp(400), unit.Dp(300)),
			app.MinSize(unit.Dp(300), unit.Dp(200)),
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

	// Initialize volume slider
	volumeSlider.Value = 0.0 // Start at maximum volume

	// Load the image
	imgFile, err := os.Open("assets/radio.png")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)
	if err != nil {
		log.Fatalf("failed to decode image: %v", err)
	}

	for e := range w.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			paint.Fill(gtx.Ops, background)

			// Handle volume changes
			if volumeSlider.Changed() {
				audioPlayer.SetVolume(1.0 - float64(volumeSlider.Value))
			}

			// Add this inside the system.FrameEvent case, before the layout code
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
					// Draw the image centered
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						imgOp := paint.NewImageOp(img)
						imgWidget := widget.Image{Src: imgOp, Scale: 1}
						return imgWidget.Layout(gtx)
					})
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
				// Add volume slider
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Create a container for the slider with some padding
					return layout.Inset{
						Left:  unit.Dp(40),
						Right: unit.Dp(40),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(th, "Volume")
								label.Color = textColor
								label.Alignment = text.Middle
								return label.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								slider := material.Slider(th, &volumeSlider, 0, 1)
								slider.Color = accent
								return slider.Layout(gtx)
							}),
						)
					})
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
