use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Clone, Deserialize, PartialEq)]
#[serde(rename_all = "camelCase")]
pub struct DaemonRequest {
    pub path: String,
    #[serde(default)]
    pub method: Option<String>,
    #[serde(default)]
    pub body: Option<Value>,
}

#[derive(Debug, Clone, Serialize, PartialEq)]
#[serde(rename_all = "camelCase")]
pub struct DaemonResponse {
    pub ok: bool,
    pub status: u16,
    pub data: Value,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
#[serde(rename_all = "camelCase")]
pub struct PlatformInfo {
    pub platform: String,
    pub version: String,
    pub packaged: bool,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
#[serde(rename_all = "camelCase")]
pub struct DaemonExitPayload {
    pub code: Option<i32>,
    pub success: bool,
}
