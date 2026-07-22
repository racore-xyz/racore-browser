use crate::error::{CommandError, CommandResult};
use crate::models::{DaemonRequest, DaemonResponse, PlatformInfo};
use crate::state::AppState;
use crate::windows::{normalize_web_url, open_browser_window};
use serde_json::{json, Value};
use tauri::{AppHandle, State};
use tauri_plugin_opener::OpenerExt;

#[tauri::command]
pub async fn daemon_status(state: State<'_, AppState>) -> CommandResult<Value> {
    Ok(state
        .daemon
        .health()
        .await
        .unwrap_or_else(|error| json!({ "ok": false, "error": error.message })))
}

#[tauri::command]
pub async fn daemon_request(
    state: State<'_, AppState>,
    request: DaemonRequest,
) -> CommandResult<DaemonResponse> {
    state.daemon.request(request).await
}

#[tauri::command]
pub fn platform_info(app: AppHandle) -> PlatformInfo {
    PlatformInfo {
        platform: std::env::consts::OS.to_owned(),
        version: app.package_info().version.to_string(),
        packaged: !cfg!(debug_assertions),
    }
}

#[tauri::command]
pub fn open_browser(app: AppHandle, url: String) -> CommandResult<bool> {
    open_browser_window(&app, &url)
}

#[tauri::command]
pub fn open_external(app: AppHandle, url: String) -> CommandResult<()> {
    let normalized = normalize_web_url(&url)?;
    app.opener()
        .open_url(normalized.as_str(), None::<&str>)
        .map_err(|error| CommandError::backend(error.to_string()))
}
