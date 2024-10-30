// src/symphonia_source.rs

use rodio::Source;
use symphonia::core::audio::{AudioBufferRef, Signal};
use symphonia::core::codecs::Decoder;
use symphonia::core::formats::FormatReader;
use std::time::Duration;
use std::thread;

pub struct SymphoniaSource {
    decoder: Box<dyn Decoder>,
    reader: Box<dyn FormatReader>,
    track_id: u32,
    current_frame_offset: usize,
    current_samples: Vec<f32>,
    channels: usize,
    source_sample_rate: u32,
    target_sample_rate: u32,
    resampling_ratio: f64,
    window_size: usize,
}

impl SymphoniaSource {
    pub fn new(
        decoder: Box<dyn Decoder>,
        reader: Box<dyn FormatReader>,
        track_id: u32,
    ) -> Self {
        let channels = decoder.codec_params().channels.unwrap().count();
        let source_sample_rate = decoder.codec_params().sample_rate.unwrap();
        let target_sample_rate = 44100;
        let resampling_ratio = target_sample_rate as f64 / source_sample_rate as f64;

        Self {
            decoder,
            reader,
            track_id,
            current_frame_offset: 0,
            current_samples: Vec::new(),
            channels,
            source_sample_rate,
            target_sample_rate,
            resampling_ratio,
            window_size: 8,
        }
    }

    fn resample(&self, input: &[f32]) -> Vec<f32> {
        if self.source_sample_rate == self.target_sample_rate {
            return input.to_vec();
        }

        let output_len = (input.len() as f64 * self.resampling_ratio) as usize;
        let mut output = Vec::with_capacity(output_len);

        for i in 0..output_len {
            let pos = i as f64 / self.resampling_ratio;
            let pos_int = pos.floor() as isize;

            for ch in 0..self.channels {
                let mut sum = 0.0;
                let mut sum_weight = 0.0;

                let window_start = -(self.window_size as isize);
                let window_end = self.window_size as isize;

                for j in window_start..=window_end {
                    let in_pos = pos_int + j;
                    if in_pos >= 0 && (in_pos as usize) < input.len() / self.channels {
                        let x = std::f64::consts::PI * (pos - in_pos as f64);
                        let mut weight = if x == 0.0 {
                            1.0
                        } else {
                            x.sin() / x
                        };

                        if j != 0 {
                            let window = (std::f64::consts::PI * j as f64 / self.window_size as f64).sin() /
                                       (std::f64::consts::PI * j as f64 / self.window_size as f64);
                            weight *= window;
                        }

                        let idx = (in_pos as usize) * self.channels + ch;
                        if idx < input.len() {
                            sum += weight as f32 * input[idx];
                            sum_weight += weight as f32;
                        }
                    }
                }

                output.push(if sum_weight > 0.0 { sum / sum_weight } else { 0.0 });
            }
        }

        output
    }

    fn next_packet(&mut self) -> Option<Vec<f32>> {
        loop {
            let packet = self.reader.next_packet().ok()?;
            
            if packet.track_id() != self.track_id {
                continue;
            }

            thread::sleep(Duration::from_millis(1));

            match self.decoder.decode(&packet) {
                Ok(decoded) => {
                    return match decoded {
                        AudioBufferRef::F32(buf) => {
                            let mut samples = Vec::with_capacity(buf.frames() * self.channels);
                            for frame in 0..buf.frames() {
                                for channel in 0..self.channels {
                                    samples.push(buf.chan(channel)[frame]);
                                }
                            }
                            Some(self.resample(&samples))
                        }
                        _ => None,
                    };
                }
                Err(_) => continue,
            }
        }
    }
}

impl Iterator for SymphoniaSource {
    type Item = f32;

    fn next(&mut self) -> Option<Self::Item> {
        if self.current_frame_offset >= self.current_samples.len() {
            self.current_samples = self.next_packet()?;
            self.current_frame_offset = 0;
        }

        let sample = self.current_samples.get(self.current_frame_offset).copied();
        self.current_frame_offset += 1;
        sample
    }
}

impl Source for SymphoniaSource {
    fn current_frame_len(&self) -> Option<usize> {
        None
    }

    fn channels(&self) -> u16 {
        self.channels as u16
    }

    fn sample_rate(&self) -> u32 {
        self.source_sample_rate
    }

    fn total_duration(&self) -> Option<Duration> {
        None
    }
}

unsafe impl Send for SymphoniaSource {}
unsafe impl Sync for SymphoniaSource {}

