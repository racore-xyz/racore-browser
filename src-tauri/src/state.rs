use crate::daemon::DaemonClient;
use crate::error::{CommandError, CommandResult};
use crate::models::DaemonExitPayload;
use std::sync::Mutex;
use std::time::Duration;
use tauri::{AppHandle, Emitter, Manager};
use tauri_plugin_shell::{
    process::{CommandChild, CommandEvent},
    ShellExt,
};

const DAEMON_EXIT_EVENT: &str = "racore://daemon-exit";

#[derive(Default)]
struct OwnedProcess {
    child: Option<CommandChild>,
    pid: Option<u32>,
}

pub struct AppState {
    pub daemon: DaemonClient,
    process: Mutex<OwnedProcess>,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            daemon: DaemonClient::new(),
            process: Mutex::new(OwnedProcess::default()),
        }
    }

    pub async fn start_daemon(&self, app: &AppHandle) -> CommandResult<()> {
        if self.daemon.health().await.is_ok() {
            return Ok(());
        }

        let resource_dir = app
            .path()
            .resource_dir()
            .map_err(|error| CommandError::backend(error.to_string()))?;
        let kubo_name = if cfg!(windows) { "ipfs.exe" } else { "ipfs" };
        let kubo_path = resource_dir.join("kubo").join(kubo_name);
        if !kubo_path.is_file() {
            return Err(CommandError::backend(format!(
                "bundled Kubo executable is missing: {}",
                kubo_path.display()
            )));
        }

        let command = app
            .shell()
            .sidecar("racored")
            .map_err(|error| CommandError::backend(error.to_string()))?
            .env("RACORE_KUBO_PATH", &kubo_path)
            .current_dir(&resource_dir);
        let (mut events, child) = command
            .spawn()
            .map_err(|error| CommandError::backend(error.to_string()))?;
        let pid = child.pid();
        {
            let mut process = self.process.lock().unwrap_or_else(|lock| lock.into_inner());
            process.pid = Some(pid);
            process.child = Some(child);
        }

        let handle = app.clone();
        tauri::async_runtime::spawn(async move {
            while let Some(event) = events.recv().await {
                if let CommandEvent::Terminated(terminated) = event {
                    let state = handle.state::<AppState>();
                    let mut process = state
                        .process
                        .lock()
                        .unwrap_or_else(|lock| lock.into_inner());
                    if process.pid == Some(pid) {
                        process.pid = None;
                        process.child.take();
                    }
                    drop(process);
                    let payload = DaemonExitPayload {
                        code: terminated.code,
                        success: terminated.code == Some(0),
                    };
                    let _ = handle.emit(DAEMON_EXIT_EVENT, payload);
                    break;
                }
            }
        });

        for _ in 0..43 {
            if self.daemon.health().await.is_ok() {
                return Ok(());
            }
            tokio::time::sleep(Duration::from_millis(350)).await;
        }
        self.stop_owned_daemon();
        Err(CommandError::backend(
            "racored did not become ready within 15 seconds",
        ))
    }

    pub fn stop_owned_daemon(&self) {
        let child = self
            .process
            .lock()
            .unwrap_or_else(|lock| lock.into_inner())
            .child
            .take();
        if let Some(child) = child {
            let _ = child.kill();
        }
    }
}
