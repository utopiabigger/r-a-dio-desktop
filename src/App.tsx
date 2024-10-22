import { useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import radioLogo from "./assets/radio.png";
import "./App.css";

// This is the main React component that defines the GUI structure
function App() {
  const [isPlaying, setIsPlaying] = useState(false);
  const [volume, setVolume] = useState(1.0);

  const togglePlayback = async () => {
    try {
      const playing = await invoke<boolean>('toggle_playback');
      setIsPlaying(playing);
    } catch (error) {
      console.error('Playback error:', error);
    }
  };

  const handleVolumeChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const newVolume = parseFloat(e.target.value);
    setVolume(newVolume);
    try {
      await invoke('set_volume', { volume: newVolume });
    } catch (error) {
      console.error('Volume error:', error);
    }
  };

  return (
    <main className="container">
      <img src={radioLogo} className="logo" alt="Radio logo" />

      <div className="row">
        <button onClick={togglePlayback}>
          {isPlaying ? 'Stop' : 'Play'}
        </button>
      </div>

      <div className="row" style={{ marginTop: '1rem' }}>
        <input
          type="range"
          min="0"
          max="1"
          step="0.1"
          value={volume}
          onChange={handleVolumeChange}
          style={{ width: '200px' }}
        />
      </div>
    </main>
  );
}

export default App;
