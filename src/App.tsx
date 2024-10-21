import { useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import "./App.css";

// This is the main React component that defines the GUI structure
function App() {
  const [greetMsg, setGreetMsg] = useState("");
  

  async function greet() {
    // Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
    setGreetMsg(await invoke("greet", { name }));
  }

  return (
    <main className="container">
      <h1>r/a/dio</h1>

      <form
        className="row"
        onSubmit={(e) => {
          e.preventDefault();
          greet();
        }}
      >
        <button type="submit">Play</button>
      </form>
      <p>{greetMsg}</p>
    </main>
  );
}

export default App;
