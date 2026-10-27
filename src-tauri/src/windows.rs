use crate::error::{CommandError, CommandResult};
use std::path::PathBuf;
use tauri::{webview::NewWindowResponse, AppHandle, Manager, WebviewUrl, WebviewWindowBuilder};
use url::Url;

pub fn normalize_web_url(raw: &str) -> CommandResult<Url> {
    let trimmed = raw.trim();
    if trimmed.is_empty() {
        return Err(CommandError::invalid("URL cannot be empty"));
    }
    let candidate = if trimmed.contains("://") {
        trimmed.to_owned()
    } else {
        format!("https://{trimmed}")
    };
    let url = Url::parse(&candidate)
        .map_err(|error| CommandError::invalid(format!("invalid URL: {error}")))?;
    if !matches!(url.scheme(), "http" | "https")
        || url.host_str().is_none()
        || !url.username().is_empty()
        || url.password().is_some()
    {
        return Err(CommandError::invalid(
            "only credential-free HTTP(S) URLs are allowed",
        ));
    }
    Ok(url)
}

/// A single, non-incognito browser profile is shared by every Racore web
/// window. The platform webview stores persistent cookies, local storage,
/// IndexedDB, cache and service-worker data here, so site sign-ins that issue
/// persistent credentials survive window and application restarts.
pub fn browser_profile_dir(app: &AppHandle) -> CommandResult<PathBuf> {
    app.path()
        .app_data_dir()
        .map(|directory| directory.join("browser-profile"))
        .map_err(|error| CommandError::backend(error.to_string()))
}

pub fn open_browser_window(app: &AppHandle, raw: &str) -> CommandResult<bool> {
    let url = normalize_web_url(raw)?;
    // Window creation must run on the event loop. A synchronous command already
    // occupies that thread, so queue the work and wait for its result here.
    let handle = app.clone();
    let (sender, receiver) = std::sync::mpsc::sync_channel(1);
    app.run_on_main_thread(move || {
        let _ = sender.send(show_browser_window(&handle, url));
    })
    .map_err(|error| CommandError::backend(error.to_string()))?;
    match receiver.recv_timeout(std::time::Duration::from_secs(10)) {
        Ok(result) => result.map(|()| true),
        Err(_) => Err(CommandError::backend(
            "browser window did not start in time",
        )),
    }
}

fn show_browser_window(app: &AppHandle, url: Url) -> CommandResult<()> {
    if let Some(window) = app.get_webview_window("browser") {
        window
            .navigate(url)
            .map_err(|error| CommandError::backend(error.to_string()))?;
        let _ = window.set_focus();
        return Ok(());
    }
    let data_directory = browser_profile_dir(app)?;
    std::fs::create_dir_all(&data_directory)
        .map_err(|error| CommandError::backend(error.to_string()))?;

    WebviewWindowBuilder::new(app, "browser", WebviewUrl::External(url))
        .title("Racore Web")
        .inner_size(1280.0, 820.0)
        .data_directory(data_directory)
        .incognito(false)
        .enable_clipboard_access()
        .zoom_hotkeys_enabled(true)
        .general_autofill_enabled(true)
        .on_navigation(|next| matches!(next.scheme(), "http" | "https"))
        // Let the webview engine create target=_blank/OAuth windows. They
        // inherit the same browser profile and therefore the same login.
        .on_new_window(|next, _| {
            if matches!(next.scheme(), "http" | "https") {
                NewWindowResponse::Allow
            } else {
                NewWindowResponse::Deny
            }
        })
        .on_document_title_changed(|window, title| {
            let title = title.trim();
            if !title.is_empty() {
                let _ = window.set_title(title);
            }
        })
        .build()
        .map_err(|error| CommandError::backend(error.to_string()))?;
    Ok(())
}

pub fn browser_control(app: &AppHandle, action: &str) -> CommandResult<bool> {
    let script = match action {
        "back" => "history.back()",
        "forward" => "history.forward()",
        "reload" => "location.reload()",
        other => {
            return Err(CommandError::invalid(format!(
                "unsupported browser action: {other}"
            )))
        }
    };
    let window = app
        .get_webview_window("browser")
        .ok_or_else(|| CommandError::invalid("no Racore browser window is open"))?;
    let script = script.to_owned();
    app.run_on_main_thread(move || {
        if let Err(error) = window.eval(&script) {
            eprintln!("racore: browser action failed: {error}");
        }
    })
    .map_err(|error| CommandError::backend(error.to_string()))?;
    Ok(true)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn normalizes_hostnames_to_https() {
        assert_eq!(
            normalize_web_url(" racore.xyz/docs ").unwrap().as_str(),
            "https://racore.xyz/docs"
        );
    }

    #[test]
    fn accepts_http_and_https_only() {
        assert!(normalize_web_url("http://127.0.0.1:47831/health").is_ok());
        assert!(normalize_web_url("https://racore.xyz").is_ok());
        assert!(normalize_web_url("file:///etc/passwd").is_err());
        assert!(normalize_web_url("javascript:alert(1)").is_err());
        assert!(normalize_web_url("https://user:secret@example.com").is_err());
    }
}
