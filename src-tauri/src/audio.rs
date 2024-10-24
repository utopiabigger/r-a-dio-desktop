use rodio::{Decoder, OutputStream, Sink};
use std::io::{BufReader, Read};
use std::sync::Arc;
use reqwest::blocking::Client;

pub struct AudioManager {
    sink: Option<Arc<Sink>>,
    _stream: Option<(OutputStream, rodio::OutputStreamHandle)>, // Keep stream alive
    is_playing: bool,
}

// Implement Send for AudioManager
unsafe impl Send for AudioManager {}
unsafe impl Sync for AudioManager {}

impl Default for AudioManager {
    fn default() -> Self {
        Self {
            sink: None,
            _stream: OutputStream::try_default().ok(),
            is_playing: false,
        }
    }
}

impl AudioManager {
    pub fn is_playing(&self) -> bool {
        self.is_playing
    }

    pub fn start(&mut self) -> Result<(), String> {
        if self.is_playing {
            return Ok(());
        }

        // Get stream handle
        let (_stream, stream_handle) = OutputStream::try_default()
            .map_err(|e| e.to_string())?;

        let sink = Arc::new(Sink::try_new(&stream_handle)
            .map_err(|e| e.to_string())?);
        
        let sink_clone = sink.clone();

        // Spawn a thread to handle the streaming
        std::thread::spawn(move || {
            // Create HTTP client and fetch stream
            if let Ok(response) = Client::new()
                .get("https://relay0.r-a-d.io/main.mp3")
                .header("Icy-MetaData", "1")
                .header("User-Agent", "r-a-dio-desktop/0.1.0")
                .send()
            {
                let mut reader = BufReader::new(response);
                let mut buffer = [0; 8192]; // 8KB buffer

                loop {
                    match reader.read(&mut buffer) {
                        Ok(bytes_read) if bytes_read > 0 => {
                            // Create a new buffer with just the bytes we read
                            let chunk = buffer[..bytes_read].to_vec();
                            
                            // Try to decode and play the chunk
                            if let Ok(decoder) = Decoder::new(std::io::Cursor::new(chunk)) {
                                sink_clone.append(decoder);
                            }
                        }
                        _ => break, // Error or EOF
                    }
                }
            }
        });

        self.sink = Some(sink);
        self._stream = Some((_stream, stream_handle));
        self.is_playing = true;

        Ok(())
    }

    pub fn stop(&mut self) -> Result<(), String> {
        if let Some(sink) = self.sink.take() {
            sink.stop();
        }
        self._stream = None;
        self.is_playing = false;
        Ok(())
    }

    pub fn set_volume(&self, volume: f32) -> Result<(), String> {
        if let Some(sink) = &self.sink {
            sink.set_volume(volume);
        }
        Ok(())
    }
}
