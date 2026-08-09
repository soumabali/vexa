use ssh_core::ffi::tauri as ssh_ffi;
use ssh_core::ssh::SshManager;
use std::sync::Mutex;

struct AppState {
    ssh_manager: Mutex<SshManager>,
}

#[tauri::command]
fn get_app_version() -> String {
    format!("vexa-desktop v{}", env!("CARGO_PKG_VERSION"))
}

#[tauri::command]
fn ssh_connect(
    state: tauri::State<AppState>,
    request: ssh_ffi::SshConnectRequest,
) -> Result<ssh_ffi::SshConnectResponse, String> {
    let manager = state.ssh_manager.lock().map_err(|e| e.to_string())?;
    ssh_ffi::ssh_connect(request, &manager)
}

#[tauri::command]
fn ssh_disconnect(
    state: tauri::State<AppState>,
    session_id: String,
) -> Result<(), String> {
    let manager = state.ssh_manager.lock().map_err(|e| e.to_string())?;
    ssh_ffi::ssh_disconnect(session_id, &manager)
}

#[tauri::command]
fn generate_key() -> String {
    ssh_ffi::generate_key()
}

#[tauri::command]
fn encrypt_aes(request: ssh_ffi::EncryptRequest) -> Result<ssh_ffi::EncryptResponse, String> {
    ssh_ffi::encrypt_aes(request)
}

#[tauri::command]
fn decrypt_aes(request: ssh_ffi::DecryptRequest) -> Result<ssh_ffi::DecryptResponse, String> {
    ssh_ffi::decrypt_aes(request)
}

pub fn run() {
    tauri::Builder::default()
        .manage(AppState {
            ssh_manager: Mutex::new(SshManager::new()),
        })
        .invoke_handler(tauri::generate_handler![
            get_app_version,
            ssh_connect,
            ssh_disconnect,
            generate_key,
            encrypt_aes,
            decrypt_aes,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
