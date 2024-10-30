use rodio::{OutputStream, Sink};
use std::sync::Arc;
use reqwest::blocking::Client;
use symphonia::core::codecs::DecoderOptions;
use symphonia::core::formats::FormatOptions;
use symphonia::core::io::MediaSourceStream;
use symphonia::core::meta::MetadataOptions;
use symphonia::core::probe::Hint;
use symphonia::default;

// Import our custom types from sibling modules
use crate::media_source::MediaSourceResponse;
use crate::symphonia_source::SymphoniaSource;

// AudioManager: Main struct responsible for handling audio playback
pub struct AudioManager {
    sink: Option<Arc<Sink>>,          // Handles actual audio output, wrapped in Arc for thread safety
    stream: Option<(OutputStream, rodio::OutputStreamHandle)>, // System audio device connection
    is_playing: bool,                 // Current playback state
}

// Mark AudioManager as safe to send between threads
unsafe impl Send for AudioManager {}
unsafe impl Sync for AudioManager {}

impl Default for AudioManager {
    // Initialize a new AudioManager with default values
    fn default() -> Self {
        Self {
            sink: None,
            stream: None,
            is_playing: false,
        }
    }
}

impl AudioManager {
    // Returns current playback state
    pub fn is_playing(&self) -> bool {
        self.is_playing
    }

    // Starts audio playback
    pub fn start(&mut self) -> Result<(), String> {
        // Don't start if already playing
        if self.is_playing {
            return Ok(());
        }

        // Initialize audio output device
        let (stream, stream_handle) = OutputStream::try_default()
            .map_err(|e| e.to_string())?;

        // Create new audio sink for playback
        let sink = Arc::new(Sink::try_new(&stream_handle)
            .map_err(|e| e.to_string())?);

        self.stream = Some((stream, stream_handle));
        let sink_clone = sink.clone();

        // Spawn background thread to handle streaming
        std::thread::spawn(move || {
            // Connect to r/a/dio stream
            if let Ok(response) = Client::new()
                .get("https://relay0.r-a-d.io/main.mp3")
                .header("Icy-MetaData", "1")
                .header("User-Agent", "r-a-dio-desktop/0.1.0")
                .send()
            {
                // Set up media source and decoder pipeline
                let media_source = Box::new(MediaSourceResponse::new(response));
                let mss = MediaSourceStream::new(media_source, Default::default());

                // Probe audio format and set up decoder
                if let Ok(probed) = default::get_probe().format(
                    &Hint::new(),
                    mss,
                    &FormatOptions::default(),
                    &MetadataOptions::default(),
                ) {
                    // Get default audio track
                    if let Some(track) = probed.format.default_track() {
                        let track_id = track.id;
                        let codec_params = track.codec_params.clone();

                        // Create decoder for audio format
                        if let Ok(decoder) = default::get_codecs().make(
                            &codec_params,
                            &DecoderOptions::default(),
                        ) {
                            // Create audio source and start playback
                            let source = SymphoniaSource::new(decoder, probed.format, track_id);
                            sink_clone.append(source);
                        } else {
                            eprintln!("Failed to create decoder");
                        }
                    } else {
                        eprintln!("No default track found");
                    }
                } else {
                    eprintln!("Failed to probe audio format");
                }
            } else {
                eprintln!("Failed to connect to radio stream");
            }
        });

        self.sink = Some(sink);
        self.is_playing = true;

        Ok(())
    }

    // Stops audio playback
    pub fn stop(&mut self) -> Result<(), String> {
        if let Some(sink) = self.sink.take() {
            sink.stop();
        }
        self.stream = None;
        self.is_playing = false;
        Ok(())
    }

    // Sets playback volume (0.0 - 1.0)
    pub fn set_volume(&self, volume: f32) -> Result<(), String> {
        if let Some(sink) = &self.sink {
            sink.set_volume(volume);
        }
        Ok(())
    }
}
