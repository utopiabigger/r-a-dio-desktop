// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
mod audio;

use audio::AudioManager;
use parking_lot::Mutex;

// Wrap our state in a mutex for thread-safe access
struct State {
    audio: Mutex<AudioManager>,
}

impl Default for State {
    fn default() -> Self {
        Self {
            audio: Mutex::new(AudioManager::default()),
        }
    }
}

#[tauri::command]
fn toggle_playback(state: tauri::State<'_, State>) -> Result<bool, String> {
    let mut audio = state.audio.lock();
    if audio.is_playing() {
        audio.stop()?;
        Ok(false)
    } else {
        audio.start()?;
        Ok(true)
    }
}

#[tauri::command]
fn set_volume(volume: f32, state: tauri::State<'_, State>) -> Result<(), String> {
    let audio = state.audio.lock();
    audio.set_volume(volume)
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(State::default())
        .invoke_handler(tauri::generate_handler![toggle_playback, set_volume])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
