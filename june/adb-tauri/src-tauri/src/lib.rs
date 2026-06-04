use tauri_plugin_shell::ShellExt;

const SIDECAR_PORT: &str = "19527";

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let sidecar = app
                .shell()
                .sidecar("adb-sidecar")
                .map_err(|e| {
                    eprintln!(
                        "adb-sidecar not found — run: npm run build:sidecar\n{e}"
                    );
                    e
                })?
                .env("ADB_SIDECAR_PORT", SIDECAR_PORT);

            sidecar.spawn().map_err(|e| {
                eprintln!("failed to spawn adb-sidecar: {e}");
                e
            })?;

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
