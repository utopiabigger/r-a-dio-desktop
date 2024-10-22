use rodio::{Decoder, OutputStream, Sink};
use std::io::Cursor;
use tokio::sync::mpsc::{self, Sender};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use reqwest::Client;
use futures_util::StreamExt;
use std::time::Duration;

#[derive(Debug, Clone)]
pub enum AudioCommand {
    Play,
    Stop,
    SetVolume(f32),
}

pub struct AudioManager {
    tx: Option<Sender<AudioCommand>>,
    is_playing: Arc<AtomicBool>,
}

impl Default for AudioManager {
    fn default() -> Self {
        Self { 
            tx: None,
            is_playing: Arc::new(AtomicBool::new(false)),
        }
    }
}

impl AudioManager {
    pub fn is_playing(&self) -> bool {
        self.is_playing.load(Ordering::SeqCst)
    }

    pub fn start(&mut self) -> Result<(), String> {
        if self.tx.is_none() {
            let (tx, mut rx) = mpsc::channel(128);
            self.tx = Some(tx.clone());
            let is_playing = self.is_playing.clone();

            std::thread::spawn(move || {
                let (_stream, stream_handle) = OutputStream::try_default().unwrap();
                let sink = Sink::try_new(&stream_handle).unwrap();
                sink.set_volume(1.0);
                
                let rt = tokio::runtime::Runtime::new().unwrap();
                rt.block_on(async {
                    while let Some(cmd) = rx.recv().await {
                        match cmd {
                            AudioCommand::Play => {
                                let client = Client::builder()
                                    .timeout(Duration::from_secs(0))
                                    .build()
                                    .unwrap();
                                
                                if let Ok(response) = client.get("https://relay0.r-a-d.io/main.mp3")
                                    .header("Icy-MetaData", "1")
                                    .header("User-Agent", "r-a-dio-desktop/0.1.0")
                                    .send()
                                    .await 
                                {
                                    let mut stream = response.bytes_stream();
                                    let mut buffer = Vec::with_capacity(8192); // Smaller initial buffer
                                    
                                    is_playing.store(true, Ordering::SeqCst);
                                    
                                    while let Some(chunk) = stream.next().await {
                                        if let Ok(data) = chunk {
                                            buffer.extend_from_slice(&data);
                                            
                                            // Try to decode and play as soon as we have enough data
                                            if buffer.len() >= 8192 {
                                                if let Ok(decoder) = Decoder::new(Cursor::new(buffer.clone())) {
                                                    sink.append(decoder);
                                                }
                                                buffer.clear();
                                                buffer.reserve(8192);
                                            }
                                        }
                                        
                                        if !is_playing.load(Ordering::SeqCst) {
                                            sink.stop();
                                            break;
                                        }

                                        // Small sleep to prevent CPU overload
                                        std::thread::sleep(Duration::from_millis(10));
                                    }
                                }
                            },
                            AudioCommand::Stop => {
                                sink.stop();
                                is_playing.store(false, Ordering::SeqCst);
                                // Break the stream receiving loop
                                break;
                            },
                            AudioCommand::SetVolume(vol) => {
                                sink.set_volume(vol);
                            }
                        }
                    }
                });
            });

            if let Some(tx) = &self.tx {
                tx.blocking_send(AudioCommand::Play)
                    .map_err(|e| e.to_string())?;
            }
        }
        Ok(())
    }

    pub fn stop(&mut self) -> Result<(), String> {
        self.is_playing.store(false, Ordering::SeqCst);
        if let Some(tx) = &self.tx {
            tx.blocking_send(AudioCommand::Stop)
                .map_err(|e| e.to_string())?;
        }
        self.tx = None;
        Ok(())
    }

    pub fn set_volume(&self, volume: f32) -> Result<(), String> {
        if let Some(tx) = &self.tx {
            tx.blocking_send(AudioCommand::SetVolume(volume))
                .map_err(|e| e.to_string())?;
        }
        Ok(())
    }
}
