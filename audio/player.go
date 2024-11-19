package audio

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Player struct {
	isPlaying  bool
	cancelFunc context.CancelFunc
	streamer   beep.StreamSeekCloser
	volume     float64
	ctrl       *effects.Volume
}

func NewPlayer() (*Player, error) {
	return &Player{
		volume: 0.5, // Default volume to 50%
	}, nil
}

func (p *Player) PlayStream(url string) error {
	if p.isPlaying {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancelFunc = cancel

	go func() {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching stream: %v", err)
			return
		}
		defer resp.Body.Close()

		streamer, format, err := mp3.Decode(resp.Body)
		if err != nil {
			log.Printf("Error decoding MP3: %v", err)
			return
		}
		p.streamer = streamer
		defer p.streamer.Close()

		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			log.Printf("Error initializing speaker: %v", err)
			return
		}

		// Create volume controller
		p.ctrl = &effects.Volume{
			Streamer: streamer,
			Base:     2,
			Volume:   40 * (p.volume - 0.5), // Initialize with the current volume setting
			Silent:   false,
		}

		p.isPlaying = true
		speaker.Play(p.ctrl)

		// Wait for context cancellation
		<-ctx.Done()
	}()

	return nil
}

func (p *Player) Stop() {
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	if p.streamer != nil {
		speaker.Clear()
		p.streamer.Close()
		p.streamer = nil
	}
	p.isPlaying = false
}

func (p *Player) IsPlaying() bool {
	return p.isPlaying
}

func (p *Player) SetVolume(vol float64) {
	p.volume = vol
	if p.ctrl != nil {
		speaker.Lock()
		// Ensure volume stays between 0 and 1
		if vol < 0 {
			vol = 0
		} else if vol > 1 {
			vol = 1
		}
		// Convert linear volume to logarithmic scale
		// When vol is 0, Volume will be -20 (minimum volume)
		// When vol is 0.5, Volume will be 0 (default volume)
		// When vol is 1, Volume will be +20 (maximum volume)
		if vol > 0 {
			p.ctrl.Volume = 40 * (vol - 0.5) // Centers at 0 dB, ranges from -20 to +20 dB
		} else {
			p.ctrl.Volume = -999 // Effectively mute
		}
		speaker.Unlock()
	}
}
