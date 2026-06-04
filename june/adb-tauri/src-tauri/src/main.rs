#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    adb_tauri_lib::run()
}
