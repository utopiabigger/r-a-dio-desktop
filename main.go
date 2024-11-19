package main

import (
	"image/color"
	"log"
	"net/http"
	"os"
	"encoding/json"
	"time"
	"image"
	"fmt"

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
	nowPlaying string
	listeners  int
	djName     string
	djImage    image.Image
	songProgress int
	songLength   int
)

// Add these structs for the API response
type RadioAPI struct {
	Main struct {
		Np        string `json:"np"`
		Listeners int    `json:"listeners"`
		DjName    string `json:"djname"`
		IsAfk     bool   `json:"isafkstream"`
		Start     int64  `json:"start_time"`    // Song start time
		End       int64  `json:"end_time"`      // Song end time
	} `json:"main"`
}

// Add this function to calculate progress
func calculateProgress(start, end int64) (int, int) {
	now := time.Now().Unix()
	total := end - start
	current := now - start
	
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	
	return int(current), int(total)
}

// Add this function to fetch the API data
func fetchRadioData() (*RadioAPI, error) {
	resp, err := http.Get("https://r-a-d.io/api")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data RadioAPI
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("r/a/dio Desktop"),
			app.Size(unit.Dp(400), unit.Dp(600)),
			app.MinSize(unit.Dp(300), unit.Dp(500)),
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
	volumeSlider.Value = 0.5 // Start at middle position (50%)

	// Load the radio logo
	imgFile, err := os.Open("assets/radio.png")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)
	if err != nil {
		log.Fatalf("failed to decode image: %v", err)
	}

	// Load the DJ image
	djImgFile, err := os.Open("assets/hanyuu.png")
	if err != nil {
		log.Fatalf("failed to open DJ image: %v", err)
	}
	defer djImgFile.Close()

	djImage, err = png.Decode(djImgFile)
	if err != nil {
		log.Fatalf("failed to decode DJ image: %v", err)
	}

	// In the run function, modify the API fetching goroutine:
	go func() {
		// Fetch immediately on start
		if data, err := fetchRadioData(); err == nil {
			nowPlaying = data.Main.Np
			listeners = data.Main.Listeners
			djName = data.Main.DjName
			songProgress, songLength = calculateProgress(data.Main.Start, data.Main.End)
			w.Invalidate()
		}

		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			if data, err := fetchRadioData(); err == nil {
				nowPlaying = data.Main.Np
				listeners = data.Main.Listeners
				djName = data.Main.DjName
				songProgress, songLength = calculateProgress(data.Main.Start, data.Main.End)
				w.Invalidate()
			}
		}
	}()

	for e := range w.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			paint.Fill(gtx.Ops, background)

			// Handle volume changes
			if volumeSlider.Changed() {
				audioPlayer.SetVolume(float64(volumeSlider.Value))
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
					// Draw the radio logo centered
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						imgOp := paint.NewImageOp(img)
						imgWidget := widget.Image{Src: imgOp, Scale: 1}
						return imgWidget.Layout(gtx)
					})
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
				// Add DJ image and name
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Create an inset to constrain the width of the DJ section
						return layout.Inset{
							Left: unit.Dp(20),
							Right: unit.Dp(20),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								// DJ Image
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									djImg := paint.NewImageOp(djImage)
									djImgWidget := widget.Image{Src: djImg, Scale: 0.8}
									return layout.Center.Layout(gtx, djImgWidget.Layout) // Center the image
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
								// DJ Name and Listeners
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											label := material.Caption(th, "DJ: " + djName)
											label.Color = textColor
											label.Alignment = text.Middle
											return layout.Center.Layout(gtx, label.Layout)
										}),
										layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											label := material.Caption(th, fmt.Sprintf("Listeners: %d", listeners))
											label.Color = textColor
											label.Alignment = text.Middle
											return layout.Center.Layout(gtx, label.Layout)
										}),
									)
								}),
							)
						})
					})
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
				// Add Now Playing text
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(th, nowPlaying)
						label.Color = textColor
						label.Alignment = text.Middle
						return label.Layout(gtx)
					})
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Create a container for the progress bar
						return layout.Inset{
							Left:  unit.Dp(40),
							Right: unit.Dp(40),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							// Format time as MM:SS
							currentTime := fmt.Sprintf("%02d:%02d", songProgress/60, songProgress%60)
							totalTime := fmt.Sprintf("%02d:%02d", songLength/60, songLength%60)
							
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								// Progress bar
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									var progress float32
									if songLength > 0 {
										progress = float32(songProgress) / float32(songLength)
									}
									progressBar := material.ProgressBar(th, progress)
									progressBar.Color = accent
									return progressBar.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
								// Time display
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											label := material.Caption(th, currentTime)
											label.Color = textColor
											return label.Layout(gtx)
										}),
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											label := material.Caption(th, totalTime)
											label.Color = textColor
											label.Alignment = text.End
											return label.Layout(gtx)
										}),
									)
								}),
							)
						})
					})
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
				// Volume slider (rest remains the same)
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
