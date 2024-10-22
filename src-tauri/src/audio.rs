use rodio::{Decoder, OutputStream, Sink};
use std::io::Cursor;
use tokio::sync::mpsc::{self, Sender};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

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
            let (tx, mut rx) = mpsc::channel(32);
            self.tx = Some(tx.clone());
            let is_playing = self.is_playing.clone();

            std::thread::spawn(move || {
                let (_stream, stream_handle) = OutputStream::try_default().unwrap();
                let sink = Sink::try_new(&stream_handle).unwrap();
                
                let rt = tokio::runtime::Runtime::new().unwrap();
                rt.block_on(async {
                    while let Some(cmd) = rx.recv().await {
                        match cmd {
                            AudioCommand::Play => {
                                if let Ok(response) = reqwest::get("https://relay0.r-a-d.io/main.mp3").await {
                                    if let Ok(bytes) = response.bytes().await {
                                        let cursor = Cursor::new(bytes);
                                        if let Ok(decoder) = Decoder::new(cursor) {
                                            sink.append(decoder);
                                            is_playing.store(true, Ordering::SeqCst);
                                        }
                                    }
                                }
                            },
                            AudioCommand::Stop => {
                                sink.stop();
                                is_playing.store(false, Ordering::SeqCst);
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
