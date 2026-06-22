use aes_gcm::{
    aead::{Aead, AeadCore, KeyInit, OsRng},
    Aes256Gcm, Key as AesKey, Nonce,
};
use argon2::{
    password_hash::SaltString,
    Argon2,
};
use chacha20poly1305::ChaCha20Poly1305;
use rand::RngCore;
use secrecy::{ExposeSecret, SecretString, SecretVec};
use zeroize::{Zeroize, ZeroizeOnDrop};

#[derive(Debug, thiserror::Error)]
pub enum CryptoError {
    #[error("encryption failed")]
    Encrypt,
    #[error("decryption failed")]
    Decrypt,
    #[error("hashing failed")]
    Hash,
    #[error("invalid input")]
    InvalidInput,
}

pub fn generate_key() -> SecretVec<u8> {
    let mut key = vec![0u8; 32];
    OsRng.fill_bytes(&mut key);
    SecretVec::new(key)
}

pub fn generate_nonce() -> Vec<u8> {
    let mut nonce = vec![0u8; 12];
    OsRng.fill_bytes(&mut nonce);
    nonce
}

pub fn encrypt(data: &[u8], key: &SecretVec<u8>) -> Result<Vec<u8>, CryptoError> {
    if data.is_empty() || key.expose_secret().len() != 32 {
        return Err(CryptoError::InvalidInput);
    }
    let aes_key = AesKey::<Aes256Gcm>::from_slice(key.expose_secret());
    let cipher = Aes256Gcm::new(aes_key);
    let nonce = Aes256Gcm::generate_nonce(&mut OsRng);
    let mut ciphertext = cipher
        .encrypt(&nonce, data)
        .map_err(|_| CryptoError::Encrypt)?;
    let mut out = nonce.to_vec();
    out.append(&mut ciphertext);
    Ok(out)
}

pub fn decrypt(data: &[u8], key: &SecretVec<u8>) -> Result<Vec<u8>, CryptoError> {
    if data.len() < 12 || key.expose_secret().len() != 32 {
        return Err(CryptoError::InvalidInput);
    }
    let (nonce, ciphertext) = data.split_at(12);
    let aes_key = AesKey::<Aes256Gcm>::from_slice(key.expose_secret());
    let cipher = Aes256Gcm::new(aes_key);
    cipher
        .decrypt(Nonce::from_slice(nonce), ciphertext)
        .map_err(|_| CryptoError::Decrypt)
}

pub fn encrypt_chacha(data: &[u8], key: &SecretVec<u8>) -> Result<Vec<u8>, CryptoError> {
    if data.is_empty() || key.expose_secret().len() != 32 {
        return Err(CryptoError::InvalidInput);
    }
    use chacha20poly1305::aead::Aead as _;
    let chacha_key =
        chacha20poly1305::Key::from_slice(key.expose_secret());
    let cipher = ChaCha20Poly1305::new(chacha_key);
    let nonce = ChaCha20Poly1305::generate_nonce(&mut OsRng);
    let mut ciphertext = cipher
        .encrypt(&nonce, data)
        .map_err(|_| CryptoError::Encrypt)?;
    let mut out = nonce.to_vec();
    out.append(&mut ciphertext);
    Ok(out)
}

pub fn decrypt_chacha(data: &[u8], key: &SecretVec<u8>) -> Result<Vec<u8>, CryptoError> {
    if data.len() < 12 || key.expose_secret().len() != 32 {
        return Err(CryptoError::InvalidInput);
    }
    use chacha20poly1305::aead::Aead as _;
    let (nonce, ciphertext) = data.split_at(12);
    let chacha_key =
        chacha20poly1305::Key::from_slice(key.expose_secret());
    let cipher = ChaCha20Poly1305::new(chacha_key);
    cipher
        .decrypt(chacha20poly1305::Nonce::from_slice(nonce), ciphertext)
        .map_err(|_| CryptoError::Decrypt)
}

pub fn hash_password(password: &SecretString, salt: &[u8]) -> Result<Vec<u8>, CryptoError> {
    let salt = SaltString::encode_b64(salt).map_err(|_| CryptoError::Hash)?;
    let argon2 = Argon2::default();
    let mut hash = [0u8; 32];
    argon2
        .hash_password_into(
            password.expose_secret().as_bytes(),
            salt.as_str().as_bytes(),
            &mut hash,
        )
        .map_err(|_| CryptoError::Hash)?;
    Ok(hash.to_vec())
}

pub fn derive_key(password: &SecretString, salt: &[u8]) -> Result<SecretVec<u8>, CryptoError> {
    let hash = hash_password(password, salt)?;
    Ok(SecretVec::new(hash))
}

pub fn secure_zero(memory: &mut [u8]) {
    memory.zeroize();
}

#[derive(Zeroize, ZeroizeOnDrop)]
struct SensitiveBuffer {
    #[zeroize(skip)]
    data: Vec<u8>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_encrypt_decrypt_aes() {
        let key = generate_key();
        let plaintext = b"hello world";
        let encrypted = encrypt(plaintext, &key).unwrap();
        assert_ne!(encrypted, plaintext);
        let decrypted = decrypt(&encrypted, &key).unwrap();
        assert_eq!(decrypted, plaintext);
    }

    #[test]
    fn test_encrypt_decrypt_chacha() {
        let key = generate_key();
        let plaintext = b"hello chacha";
        let encrypted = encrypt_chacha(plaintext, &key).unwrap();
        assert_ne!(encrypted, plaintext);
        let decrypted = decrypt_chacha(&encrypted, &key).unwrap();
        assert_eq!(decrypted, plaintext);
    }

    #[test]
    fn test_derive_key() {
        let password = SecretString::new(String::from("my_password"));
        let salt = b"random_salt_here";
        let key1 = derive_key(&password, salt).unwrap();
        let key2 = derive_key(&password, salt).unwrap();
        assert_eq!(key1.expose_secret(), key2.expose_secret());
    }

    #[test]
    fn test_secure_zero() {
        let mut buf = vec![1u8; 32];
        secure_zero(&mut buf);
        assert!(buf.iter().all(|&b| b == 0));
    }

    #[test]
    fn test_invalid_key_length() {
        let bad_key = SecretVec::new(vec![0u8; 16]);
        let res = encrypt(b"data", &bad_key);
        assert!(matches!(res, Err(CryptoError::InvalidInput)));
    }
}
