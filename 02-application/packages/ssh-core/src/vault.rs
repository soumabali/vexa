use crate::crypto::{decrypt, encrypt, derive_key, generate_key, secure_zero};
use crate::crypto::CryptoError;
use rand::RngCore;
use secrecy::{ExposeSecret, SecretString, SecretVec};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, thiserror::Error)]
pub enum VaultError {
    #[error("crypto error")]
    Crypto(#[from] CryptoError),
    #[error("invalid password")]
    InvalidPassword,
    #[error("credential not found")]
    NotFound,
    #[error("serialization error")]
    Serialize,
}

#[derive(Serialize, Deserialize, Zeroize, ZeroizeOnDrop)]
pub struct Credential {
    pub id: String,
    pub name: String,
    pub username: String,
    #[zeroize(skip)]
    pub encrypted_password: Vec<u8>,
    #[zeroize(skip)]
    pub encrypted_private_key: Option<Vec<u8>>,
}

#[derive(Serialize, Deserialize, Clone)]
pub struct VaultData {
    salt: Vec<u8>,
    encrypted_master_key: Vec<u8>,
    credentials: Vec<Vec<u8>>,
}

pub struct Vault {
    data: VaultData,
    master_key: Option<SecretVec<u8>>,
}

impl Vault {
    pub fn create(password: &SecretString) -> Result<Vault, VaultError> {
        let mut salt = vec![0u8; 32];
        rand::thread_rng().fill_bytes(&mut salt);
        let master_key = generate_key();
        let derived = derive_key(password, &salt)?;
        let encrypted_master_key = encrypt(master_key.expose_secret(), &derived)?;

        Ok(Vault {
            data: VaultData {
                salt,
                encrypted_master_key,
                credentials: Vec::new(),
            },
            master_key: Some(master_key),
        })
    }

    pub fn unlock(data: VaultData, password: &SecretString) -> Result<Vault, VaultError> {
        let derived = derive_key(password, &data.salt)?;
        let master_key = decrypt(&data.encrypted_master_key, &derived)
            .map_err(|_| VaultError::InvalidPassword)?;
        Ok(Vault {
            data,
            master_key: Some(SecretVec::new(master_key)),
        })
    }

    pub fn store_credential(
        &mut self,
        credential: Credential,
    ) -> Result<String, VaultError> {
        let master_key = self.master_key.as_ref().ok_or(VaultError::InvalidPassword)?;
        let cred_json = serde_json::to_vec(&credential).map_err(|_| VaultError::Serialize)?;
        let encrypted = encrypt(&cred_json, master_key)?;
        self.data.credentials.push(encrypted);
        Ok(credential.id.clone())
    }

    pub fn retrieve_credential(&self,
        id: &str,
    ) -> Result<Credential, VaultError> {
        let master_key = self.master_key.as_ref().ok_or(VaultError::InvalidPassword)?;
        for encrypted in &self.data.credentials {
            let decrypted = decrypt(encrypted, master_key)?;
            let cred: Credential =
                serde_json::from_slice(&decrypted).map_err(|_| VaultError::Serialize)?;
            if cred.id == id {
                return Ok(cred);
            }
        }
        Err(VaultError::NotFound)
    }

    pub fn change_password(
        &mut self,
        old_password: &SecretString,
        new_password: &SecretString,
    ) -> Result<(), VaultError> {
        // Verify old password
        let derived_old = derive_key(old_password, &self.data.salt)?;
        let master_key = decrypt(&self.data.encrypted_master_key, &derived_old)
            .map_err(|_| VaultError::InvalidPassword)?;

        // Decrypt all credentials with old key
        let mut credentials = Vec::new();
        for encrypted in &self.data.credentials {
            let decrypted = decrypt(encrypted, &SecretVec::new(master_key.clone()))?;
            credentials.push(decrypted);
        }

        // Generate new salt and derive new key
        let mut salt = vec![0u8; 32];
        rand::thread_rng().fill_bytes(&mut salt);
        let new_derived = derive_key(new_password, &salt)?;
        let new_master = generate_key();
        let encrypted_master = encrypt(new_master.expose_secret(), &new_derived)?;

        // Re-encrypt all credentials
        let mut new_credentials = Vec::new();
        for cred_data in &credentials {
            let encrypted = encrypt(cred_data, &new_master)?;
            new_credentials.push(encrypted);
        }

        // Securely clear old master key memory
        let mut old_key = master_key;
        secure_zero(&mut old_key);

        self.data = VaultData {
            salt,
            encrypted_master_key: encrypted_master,
            credentials: new_credentials,
        };
        self.master_key = Some(new_master);
        Ok(())
    }

    pub fn data(&self) -> &VaultData {
        &self.data
    }

    pub fn is_unlocked(&self) -> bool {
        self.master_key.is_some()
    }
}

impl Drop for Vault {
    fn drop(&mut self) {
        if let Some(mut key) = self.master_key.take() {
            let mut secret = key.expose_secret().clone();
            secure_zero(&mut secret);
        }
    }
}

use zeroize::{Zeroize, ZeroizeOnDrop};

#[cfg(test)]
mod tests {
    use super::*;

    fn test_password() -> SecretString {
        SecretString::new(String::from("test_password_123"))
    }

    #[test]
    fn test_create_and_unlock() {
        let password = test_password();
        let vault = Vault::create(&password).unwrap();
        assert!(vault.is_unlocked());
        let data = vault.data().clone();
        let unlocked = Vault::unlock(data, &password).unwrap();
        assert!(unlocked.is_unlocked());
    }

    #[test]
    fn test_wrong_password() {
        let password = test_password();
        let vault = Vault::create(&password).unwrap();
        let data = vault.data().clone();
        let wrong = SecretString::new(String::from("wrong"));
        assert!(Vault::unlock(data, &wrong).is_err());
    }

    #[test]
    fn test_store_and_retrieve() {
        let password = test_password();
        let mut vault = Vault::create(&password).unwrap();
        let credential = Credential {
            id: Uuid::new_v4().to_string(),
            name: "test".to_string(),
            username: "user".to_string(),
            encrypted_password: vec![1, 2, 3],
            encrypted_private_key: None,
        };
        let id = credential.id.clone();
        vault.store_credential(credential).unwrap();
        let retrieved = vault.retrieve_credential(&id).unwrap();
        assert_eq!(retrieved.username, "user");
    }

    #[test]
    fn test_change_password() {
        let old_password = test_password();
        let mut vault = Vault::create(&old_password).unwrap();
        let credential = Credential {
            id: Uuid::new_v4().to_string(),
            name: "test".to_string(),
            username: "user".to_string(),
            encrypted_password: vec![1, 2, 3],
            encrypted_private_key: None,
        };
        let id = credential.id.clone();
        vault.store_credential(credential).unwrap();

        let new_password = SecretString::new(String::from("new_password_456"));
        vault.change_password(&old_password, &new_password).unwrap();

        let data = vault.data().clone();
        let unlocked = Vault::unlock(data, &new_password).unwrap();
        let retrieved = unlocked.retrieve_credential(&id).unwrap();
        assert_eq!(retrieved.username, "user");
    }
}
