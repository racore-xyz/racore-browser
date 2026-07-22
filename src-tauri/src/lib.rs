mod commands;
mod daemon;
mod error;
mod models;
mod state;
mod windows;

use state::AppState;
use tauri::{Manager, RunEvent};

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let app = tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(
            tauri_plugin_opener::Builder::new()
                .open_js_links_on_click(false)
                .build(),
        )
        .manage(AppState::new())
        .invoke_handler(tauri::generate_handler![
            commands::daemon_status,
            commands::daemon_request,
            commands::platform_info,
            commands::open_browser,
            commands::open_external,
        ])
        .setup(|app| {
            let handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                if let Err(error) = handle.state::<AppState>().start_daemon(&handle).await {
                    eprintln!("failed to start racored: {error}");
                }
            });
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("failed to build Racore Browser");

    app.run(|handle, event| {
        if matches!(event, RunEvent::Exit) {
            handle.state::<AppState>().stop_owned_daemon();
        }
    });
}
