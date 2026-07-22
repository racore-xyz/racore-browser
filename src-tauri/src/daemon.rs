use crate::error::{CommandError, CommandResult};
use crate::models::{DaemonRequest, DaemonResponse};
use percent_encoding::percent_decode_str;
use reqwest::{Client, Method};
use serde_json::{json, Value};
use std::time::Duration;

const DAEMON_ORIGIN: &str = "http://127.0.0.1:47831";
const MAX_RESPONSE_BYTES: usize = 8 * 1024 * 1024;

pub struct DaemonClient {
    client: Client,
}

impl DaemonClient {
    pub fn new() -> Self {
        let client = Client::builder()
            .connect_timeout(Duration::from_secs(2))
            .timeout(Duration::from_secs(60))
            .build()
            .expect("static daemon HTTP client configuration must be valid");
        Self { client }
    }

    pub async fn health(&self) -> CommandResult<Value> {
        let response = self
            .client
            .get(format!("{DAEMON_ORIGIN}/health"))
            .timeout(Duration::from_millis(900))
            .send()
            .await
            .map_err(|error| CommandError::backend(error.to_string()))?;

        if !response.status().is_success() {
            return Err(CommandError::backend(format!(
                "racored health returned {}",
                response.status().as_u16()
            )));
        }
        response
            .json()
            .await
            .map_err(|error| CommandError::backend(error.to_string()))
    }

    pub async fn request(&self, request: DaemonRequest) -> CommandResult<DaemonResponse> {
        let validated = ValidatedRequest::try_from(request)?;
        let url = format!("{DAEMON_ORIGIN}{}", validated.path);
        let mut pending = self.client.request(validated.method, url);
        if let Some(body) = validated.body {
            pending = pending.json(&body);
        }

        let response = pending
            .send()
            .await
            .map_err(|error| CommandError::backend(error.to_string()))?;
        let status = response.status();
        if response
            .content_length()
            .is_some_and(|size| size > MAX_RESPONSE_BYTES as u64)
        {
            return Err(CommandError::backend("racored response exceeded 8 MiB"));
        }
        let bytes = response
            .bytes()
            .await
            .map_err(|error| CommandError::backend(error.to_string()))?;
        if bytes.len() > MAX_RESPONSE_BYTES {
            return Err(CommandError::backend("racored response exceeded 8 MiB"));
        }

        let data = serde_json::from_slice(&bytes)
            .unwrap_or_else(|_| json!({ "text": String::from_utf8_lossy(&bytes).into_owned() }));
        Ok(DaemonResponse {
            ok: status.is_success(),
            status: status.as_u16(),
            data,
        })
    }
}

struct ValidatedRequest {
    path: String,
    method: Method,
    body: Option<Value>,
}

impl TryFrom<DaemonRequest> for ValidatedRequest {
    type Error = CommandError;

    fn try_from(request: DaemonRequest) -> Result<Self, Self::Error> {
        validate_path(&request.path)?;
        let method = request
            .method
            .as_deref()
            .unwrap_or("GET")
            .to_ascii_uppercase()
            .parse::<Method>()
            .map_err(|_| CommandError::invalid("unsupported daemon method"))?;
        if !matches!(
            method,
            Method::GET | Method::POST | Method::PUT | Method::DELETE
        ) {
            return Err(CommandError::invalid("unsupported daemon method"));
        }
        if !route_allows(&method, request.path.split('?').next().unwrap_or("")) {
            return Err(CommandError::invalid(
                "daemon method and path are not in the Racore API allowlist",
            ));
        }
        if matches!(method, Method::GET | Method::DELETE) && request.body.is_some() {
            return Err(CommandError::invalid(
                "GET and DELETE daemon requests cannot contain a body",
            ));
        }

        Ok(Self {
            path: request.path,
            method,
            body: request.body,
        })
    }
}

fn validate_path(path: &str) -> CommandResult<()> {
    if !path.starts_with('/') || path.starts_with("//") {
        return Err(CommandError::invalid(
            "daemon path must be an absolute local path",
        ));
    }
    if path.contains(['\\', '#', '\0']) {
        return Err(CommandError::invalid(
            "daemon path contains unsafe characters",
        ));
    }
    let path_only = path.split('?').next().unwrap_or(path);
    let decoded = percent_decode_str(path_only)
        .decode_utf8()
        .map_err(|_| CommandError::invalid("daemon path is not valid UTF-8"))?;
    if decoded.contains('\\')
        || decoded
            .split('/')
            .any(|segment| matches!(segment, "." | ".."))
    {
        return Err(CommandError::invalid(
            "daemon path traversal is not allowed",
        ));
    }
    Ok(())
}

fn route_allows(method: &Method, path: &str) -> bool {
    let segments: Vec<_> = path.trim_matches('/').split('/').collect();
    match (method, segments.as_slice()) {
        (&Method::GET, ["health"])
        | (&Method::GET, ["v1", "providers"])
        | (&Method::POST, ["v1", "chat"])
        | (&Method::GET, ["v1", "mesh", "status"])
        | (&Method::GET, ["v1", "mesh", "peers"])
        | (&Method::POST, ["v1", "mesh", "broadcast"])
        | (&Method::GET, ["v1", "ipfs", "status"])
        | (&Method::GET, ["v1", "authority", "domains"])
        | (&Method::POST, ["v1", "authority", "domains"]) => true,
        (&Method::DELETE, ["v1", "providers", provider]) => safe_identifier(provider),
        (&Method::PUT, ["v1", "providers", provider, "connect"])
        | (&Method::GET, ["v1", "providers", provider, "health"]) => safe_identifier(provider),
        (&Method::GET, ["v1", "authority", "domains", domain, "available"]) => safe_domain(domain),
        _ => false,
    }
}

fn safe_identifier(value: &str) -> bool {
    !value.is_empty()
        && value.len() <= 64
        && value
            .bytes()
            .all(|byte| byte.is_ascii_alphanumeric() || matches!(byte, b'-' | b'_'))
}

fn safe_domain(value: &str) -> bool {
    !value.is_empty()
        && value.len() <= 253
        && value
            .bytes()
            .all(|byte| byte.is_ascii_alphanumeric() || matches!(byte, b'-' | b'.'))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn request(path: &str, method: &str, body: Option<Value>) -> DaemonRequest {
        DaemonRequest {
            path: path.into(),
            method: Some(method.into()),
            body,
        }
    }

    #[test]
    fn allows_current_react_routes() {
        let routes = [
            ("/health", "GET"),
            ("/v1/providers", "GET"),
            ("/v1/providers/open-router/connect", "PUT"),
            ("/v1/providers/open-router", "DELETE"),
            ("/v1/providers/open-router/health", "GET"),
            ("/v1/chat", "POST"),
            ("/v1/mesh/status", "GET"),
            ("/v1/mesh/peers", "GET"),
            ("/v1/mesh/broadcast", "POST"),
            ("/v1/ipfs/status", "GET"),
            ("/v1/authority/domains", "GET"),
            ("/v1/authority/domains/example.com/available", "GET"),
            ("/v1/authority/domains", "POST"),
        ];
        for (path, method) in routes {
            let body = matches!(method, "POST" | "PUT").then(|| json!({}));
            assert!(ValidatedRequest::try_from(request(path, method, body)).is_ok());
        }
    }

    #[test]
    fn rejects_origin_escape_and_path_traversal() {
        for path in [
            "https://example.com/health",
            "//example.com/health",
            "/v1/../health",
            "/v1/%2e%2e/health",
            "/v1/%5c%5cexample.com",
            "/health#fragment",
        ] {
            assert!(ValidatedRequest::try_from(request(path, "GET", None)).is_err());
        }
    }

    #[test]
    fn rejects_unlisted_routes_methods_and_invalid_bodies() {
        assert!(ValidatedRequest::try_from(request("/debug", "GET", None)).is_err());
        assert!(ValidatedRequest::try_from(request("/health", "PATCH", None)).is_err());
        assert!(ValidatedRequest::try_from(request(
            "/health",
            "GET",
            Some(json!({ "unexpected": true }))
        ))
        .is_err());
    }
}
