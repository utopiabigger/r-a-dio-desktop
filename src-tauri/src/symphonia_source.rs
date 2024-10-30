// src/symphonia_source.rs

use rodio::Source;
use symphonia::core::audio::SampleBuffer;
use symphonia::core::codecs::Decoder as SymphoniaDecoder;
use symphonia::core::errors::Error as SymphoniaError;
use symphonia::core::formats::FormatReader;
use std::time::Duration;

// SymphoniaSource: Handles audio decoding and processing
pub struct SymphoniaSource {
    decoder: Box<dyn SymphoniaDecoder + Send>,    // Audio decoder with Send bound
    reader: Box<dyn FormatReader + Send>,         // Format reader with Send bound
    track_id: u32,                         // Audio track identifier
    sample_buffer: Option<SampleBuffer<f32>>, // Buffer for decoded audio
    current_sample_index: usize,           // Current position in buffer
}

// Explicitly implement Send and Sync
unsafe impl Send for SymphoniaSource {}
unsafe impl Sync for SymphoniaSource {}

impl SymphoniaSource {
    // Creates new audio source with decoder and format reader
    pub fn new(
        decoder: Box<dyn SymphoniaDecoder + Send>,
        reader: Box<dyn FormatReader + Send>,
        track_id: u32,
    ) -> Self {
        Self {
            decoder,
            reader,
            track_id,
            sample_buffer: None,
            current_sample_index: 0,
        }
    }
}

// Implement Iterator to process audio samples
impl Iterator for SymphoniaSource {
    type Item = f32;

    fn next(&mut self) -> Option<Self::Item> {
        loop {
            // Return buffered samples if available
            if let Some(ref buf) = self.sample_buffer {
                if self.current_sample_index < buf.len() {
                    let sample = buf.samples()[self.current_sample_index];
                    self.current_sample_index += 1;
                    return Some(sample);
                }
                // Reset buffer when consumed
                self.sample_buffer = None;
                self.current_sample_index = 0;
            }

            // Read next packet of audio data
            let packet = match self.reader.next_packet() {
                Ok(packet) => packet,
                Err(_) => return None,
            };

            // Skip packets not belonging to our audio track
            if packet.track_id() != self.track_id {
                continue;
            }

            // Decode audio packet
            match self.decoder.decode(&packet) {
                Ok(decoded) => {
                    // Create sample buffer from decoded audio
                    let mut sample_buf = SampleBuffer::<f32>::new(
                        decoded.capacity() as u64,
                        *decoded.spec(),
                    );
                    sample_buf.copy_interleaved_ref(decoded);
                    self.sample_buffer = Some(sample_buf);
                    self.current_sample_index = 0;
                }
                Err(SymphoniaError::DecodeError(_)) => continue,
                Err(_) => return None,
            }
        }
    }
}

// Implement Source trait for rodio audio playback
impl Source for SymphoniaSource {
    fn current_frame_len(&self) -> Option<usize> {
        None    // Unknown frame length for streaming
    }

    fn channels(&self) -> u16 {
        self.decoder.codec_params().channels.unwrap().count() as u16
    }

    fn sample_rate(&self) -> u32 {
        self.decoder.codec_params().sample_rate.unwrap()
    }

    fn total_duration(&self) -> Option<Duration> {
        None    // Unknown duration for streaming
    }
}
