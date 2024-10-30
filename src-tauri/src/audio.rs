use rodio::{OutputStream, Sink};
use std::sync::Arc;
use reqwest::blocking::Client;
use symphonia::core::codecs::DecoderOptions;
use symphonia::core::formats::FormatOptions;
use symphonia::core::io::{MediaSourceStream, MediaSourceStreamOptions};
use symphonia::core::meta::MetadataOptions;
use symphonia::core::probe::Hint;
use symphonia::default;

#[cfg(windows)]
use windows::Win32::System::Threading::{
    GetCurrentThread, SetThreadPriority,
    THREAD_PRIORITY_TIME_CRITICAL,
};

use crate::media_source::MediaSourceResponse;
use crate::symphonia_source::SymphoniaSource;

pub struct AudioManager {
    sink: Option<Arc<Sink>>,
    _stream: Option<OutputStream>,
    is_playing: bool,
}

impl Default for AudioManager {
    fn default() -> Self {
        Self {
            sink: None,
            _stream: None,
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

        let (stream, stream_handle) = OutputStream::try_default()
            .map_err(|e| e.to_string())?;

        let sink = Sink::try_new(&stream_handle)
            .map_err(|e| e.to_string())?;
        let sink = Arc::new(sink);
        let sink_clone = sink.clone();

        std::thread::Builder::new()
            .name("Audio Thread".into())
            .spawn(move || {
                #[cfg(windows)]
                unsafe {
                    let thread_handle = GetCurrentThread();
                    let _ = SetThreadPriority(thread_handle, THREAD_PRIORITY_TIME_CRITICAL);
                }

                if let Ok(response) = Client::new()
                    .get("https://relay0.r-a-d.io/main.mp3")
                    .header("Icy-MetaData", "1")
                    .send()
                {
                    let media_source = MediaSourceResponse::new(response);
                    let stream_options = MediaSourceStreamOptions {
                        buffer_len: 512 * 1024,
                        ..Default::default()
                    };
                    let mss = MediaSourceStream::new(
                        Box::new(media_source),
                        stream_options
                    );

                    if let Ok(probed) = default::get_probe().format(
                        &Hint::new(),
                        mss,
                        &FormatOptions::default(),
                        &MetadataOptions::default(),
                    ) {
                        if let Some(track) = probed.format.default_track() {
                            let track_id = track.id;
                            let codec_params = track.codec_params.clone();

                            if let Ok(decoder) = default::get_codecs().make(
                                &codec_params,
                                &DecoderOptions::default(),
                            ) {
                                let source = SymphoniaSource::new(
                                    decoder, 
                                    probed.format,
                                    track_id
                                );
                                sink_clone.append(source);
                            }
                        }
                    }
                }
            })
            .map_err(|e| e.to_string())?;

        self._stream = Some(stream);
        self.sink = Some(sink);
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

unsafe impl Send for AudioManager {}
unsafe impl Sync for AudioManager {}
