package audio

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Player struct {
	isPlaying  bool
	cancelFunc context.CancelFunc
	streamer   beep.StreamSeekCloser
}

func NewPlayer() (*Player, error) {
	return &Player{}, nil
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

		p.isPlaying = true
		speaker.Play(streamer)

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