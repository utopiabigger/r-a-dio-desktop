// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]


// The main entry point for the Rust backend
fn main() {
    r_a_dio_desktop_lib::run()
}
