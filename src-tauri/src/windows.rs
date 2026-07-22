use crate::error::{CommandError, CommandResult};
use std::sync::atomic::{AtomicU64, Ordering};
use tauri::{webview::NewWindowResponse, AppHandle, Manager, WebviewUrl, WebviewWindowBuilder};
use url::Url;

static NEXT_BROWSER_WINDOW: AtomicU64 = AtomicU64::new(1);

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

pub fn open_browser_window(app: &AppHandle, raw: &str) -> CommandResult<bool> {
    let url = normalize_web_url(raw)?;
    let label = format!(
        "browser-{}",
        NEXT_BROWSER_WINDOW.fetch_add(1, Ordering::Relaxed)
    );
    let data_directory = app
        .path()
        .app_data_dir()
        .map_err(|error| CommandError::backend(error.to_string()))?
        .join("webview-racore");
    std::fs::create_dir_all(&data_directory)
        .map_err(|error| CommandError::backend(error.to_string()))?;

    WebviewWindowBuilder::new(app, label, WebviewUrl::External(url))
        .title("Racore Web")
        .inner_size(1280.0, 820.0)
        .data_directory(data_directory)
        .on_navigation(|next| matches!(next.scheme(), "http" | "https"))
        .on_new_window(|_, _| NewWindowResponse::Deny)
        .build()
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
