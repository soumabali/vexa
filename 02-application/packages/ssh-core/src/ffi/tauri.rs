use crate::crypto;
use crate::ssh::SshManager;
use crate::vault::{Vault, VaultData};
use secrecy::{ExposeSecret, SecretString, SecretVec};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct EncryptRequest {
    pub data: String,
    pub key: String,
}

#[derive(Serialize, Deserialize)]
pub struct EncryptResponse {
    pub encrypted: String,
}

#[derive(Serialize, Deserialize)]
pub struct DecryptRequest {
    pub data: String,
    pub key: String,
}

#[derive(Serialize, Deserialize)]
pub struct DecryptResponse {
    pub decrypted: String,
}

#[derive(Serialize, Deserialize)]
pub struct SshConnectRequest {
    pub host: String,
    pub port: u16,
    pub username: String,
    pub password: String,
}

#[derive(Serialize, Deserialize)]
pub struct SshConnectResponse {
    pub session_id: String,
}

#[derive(Serialize, Deserialize)]
pub struct VaultCreateRequest {
    pub password: String,
}

#[derive(Serialize, Deserialize)]
pub struct VaultUnlockRequest {
    pub data: String,
    pub password: String,
}

#[derive(Serialize, Deserialize)]
pub struct VaultResponse {
    pub success: bool,
    pub data: Option<String>,
    pub error: Option<String>,
}

pub fn encrypt_aes(request: EncryptRequest) -> Result<EncryptResponse, String> {
    let key_bytes = hex::decode(&request.key).map_err(|_| "invalid key")?;
    let key = SecretVec::new(key_bytes);
    let data = request.data.as_bytes();
    let encrypted = crypto::encrypt(data, &key).map_err(|_| "encryption failed")?;
    Ok(EncryptResponse {
        encrypted: hex::encode(encrypted),
    })
}

pub fn decrypt_aes(request: DecryptRequest) -> Result<DecryptResponse, String> {
    let key_bytes = hex::decode(&request.key).map_err(|_| "invalid key")?;
    let key = SecretVec::new(key_bytes);
    let data = hex::decode(&request.data).map_err(|_| "invalid data")?;
    let decrypted = crypto::decrypt(&data, &key).map_err(|_| "decryption failed")?;
    Ok(DecryptResponse {
        decrypted: String::from_utf8_lossy(&decrypted).to_string(),
    })
}

pub fn generate_key() -> String {
    let key = crypto::generate_key();
    hex::encode(key.expose_secret())
}

pub fn ssh_connect(
    request: SshConnectRequest,
    manager: &SshManager,
) -> Result<SshConnectResponse, String> {
    let password = SecretString::new(request.password);
    let session_id = manager
        .connect(&request.host, request.port, &request.username, &password)
        .map_err(|_| "ssh connection failed")?;
    Ok(SshConnectResponse { session_id })
}

pub fn ssh_disconnect(
    session_id: String,
    manager: &SshManager,
) -> Result<(), String> {
    Ok(manager.disconnect(&session_id).map_err(|e| e.to_string())?)
}

pub fn create_vault(request: VaultCreateRequest) -> Result<VaultResponse, String> {
    let password = SecretString::new(request.password);
    let vault = Vault::create(&password).map_err(|_| "vault creation failed")?;
    let data = serde_json::to_string(vault.data()).map_err(|_| "serialization failed")?;
    Ok(VaultResponse {
        success: true,
        data: Some(data),
        error: None,
    })
}

pub fn unlock_vault(request: VaultUnlockRequest) -> Result<VaultResponse, String> {
    let vault_data: VaultData =
        serde_json::from_str(&request.data).map_err(|_| "invalid vault data")?;
    let password = SecretString::new(request.password);
    let vault = Vault::unlock(vault_data, &password).map_err(|_| "unlock failed")?;
    let data = serde_json::to_string(vault.data()).map_err(|_| "serialization failed")?;
    Ok(VaultResponse {
        success: true,
        data: Some(data),
        error: None,
    })
}
