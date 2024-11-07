package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// RadioAPI represents the response from r/a/d.io API
type RadioAPI struct {
	Main struct {
		Np        string `json:"np"`        // Now Playing
		Listeners int    `json:"listeners"`
		DJName    string `json:"djname"`
	} `json:"main"`
}

func main() {
	log.Println("Starting application...")
	
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Application panicked: %v", r)
		}
	}()
	
	myApp := app.New()
	if myApp == nil {
		log.Fatal("Failed to create application")
	}
	
	log.Println("Creating window...")
	window := myApp.NewWindow("r/a/d.io Desktop")
	if window == nil {
		log.Fatal("Failed to create window")
	}

	log.Println("Creating UI elements...")
	// Create UI elements
	nowPlaying := widget.NewLabel("Loading...")
	listeners := widget.NewLabel("Listeners: --")
	djName := widget.NewLabel("DJ: --")
	
	// Create layout
	content := container.NewVBox(
		nowPlaying,
		listeners,
		djName,
		widget.NewButton("Play/Pause", func() {
			log.Println("Button clicked") // Add this to verify button works
		}),
	)

	// Update info every 5 seconds
	go func() {
		log.Println("Starting update goroutine...")
		for range time.Tick(5 * time.Second) {
			updateRadioInfo(nowPlaying, listeners, djName)
		}
	}()

	log.Println("Setting window content...")
	window.SetContent(content)
	window.Resize(fyne.NewSize(300, 200))
	
	log.Println("Showing window...")
	window.ShowAndRun()
}

func updateRadioInfo(nowPlaying, listeners, djName *widget.Label) {
	log.Println("Updating radio info...")
	resp, err := http.Get("https://r-a-d.io/api")
	if err != nil {
		log.Println("Error fetching API:", err)
		return
	}
	defer resp.Body.Close()

	var api RadioAPI
	if err := json.NewDecoder(resp.Body).Decode(&api); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	nowPlaying.SetText(api.Main.Np)
	listeners.SetText(fmt.Sprintf("Listeners: %d", api.Main.Listeners))
	djName.SetText(fmt.Sprintf("DJ: %s", api.Main.DJName))
} 