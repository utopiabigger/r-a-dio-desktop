mod audio;
mod media_source;
mod symphonia_source;

use std::sync::Mutex;
use tauri::State;

// Commands that will be called from the frontend
#[tauri::command]
async fn toggle_playback<'a>(audio_manager: State<'a, Mutex<audio::AudioManager>>) -> Result<bool, String> {
    let mut manager = audio_manager.lock().unwrap();
    if manager.is_playing() {
        manager.stop()?;
        Ok(false)
    } else {
        manager.start()?;
        Ok(true)
    }
}

#[tauri::command]
async fn set_volume<'a>(volume: f32, audio_manager: State<'a, Mutex<audio::AudioManager>>) -> Result<(), String> {
    let manager = audio_manager.lock().unwrap();
    manager.set_volume(volume)
}

fn main() {
    tauri::Builder::default()
        .manage(Mutex::new(audio::AudioManager::default()))
        .invoke_handler(tauri::generate_handler![toggle_playback, set_volume])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
