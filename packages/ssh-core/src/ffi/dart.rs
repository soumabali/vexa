use crate::crypto;
use crate::vault::{Vault, VaultData};
use libc::c_char;
use secrecy::{ExposeSecret, SecretString, SecretVec};
use std::ffi::{CStr, CString};

#[no_mangle]
pub extern "C" fn ssh_core_generate_key() -> *mut c_char {
    let key = crypto::generate_key();
    let hex = hex::encode(key.expose_secret());
    CString::new(hex).unwrap_or_default().into_raw()
}

#[no_mangle]
pub extern "C" fn ssh_core_free_string(s: *mut c_char) {
    if s.is_null() {
        return;
    }
    unsafe {
        let _ = CString::from_raw(s);
    }
}

#[no_mangle]
pub extern "C" fn ssh_core_encrypt_aes(
    data: *const u8,
    data_len: usize,
    key: *const c_char,
    out_len: *mut usize,
) -> *mut u8 {
    if data.is_null() || key.is_null() || out_len.is_null() {
        return std::ptr::null_mut();
    }

    let data_slice = unsafe { std::slice::from_raw_parts(data, data_len) };
    let key_str = unsafe {
        match CStr::from_ptr(key).to_str() {
            Ok(s) => s,
            Err(_) => return std::ptr::null_mut(),
        }
    };

    let key_bytes = match hex::decode(key_str) {
        Ok(b) => b,
        Err(_) => return std::ptr::null_mut(),
    };
    let secret_key = SecretVec::new(key_bytes);

    let encrypted = match crypto::encrypt(data_slice, &secret_key) {
        Ok(e) => e,
        Err(_) => return std::ptr::null_mut(),
    };

    let mut boxed = encrypted.into_boxed_slice();
    let ptr = boxed.as_mut_ptr();
    unsafe {
        *out_len = boxed.len();
    }
    std::mem::forget(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn ssh_core_decrypt_aes(
    data: *const u8,
    data_len: usize,
    key: *const c_char,
    out_len: *mut usize,
) -> *mut u8 {
    if data.is_null() || key.is_null() || out_len.is_null() {
        return std::ptr::null_mut();
    }

    let data_slice = unsafe { std::slice::from_raw_parts(data, data_len) };
    let key_str = unsafe {
        match CStr::from_ptr(key).to_str() {
            Ok(s) => s,
            Err(_) => return std::ptr::null_mut(),
        }
    };

    let key_bytes = match hex::decode(key_str) {
        Ok(b) => b,
        Err(_) => return std::ptr::null_mut(),
    };
    let secret_key = SecretVec::new(key_bytes);

    let decrypted = match crypto::decrypt(data_slice, &secret_key) {
        Ok(d) => d,
        Err(_) => return std::ptr::null_mut(),
    };

    let mut boxed = decrypted.into_boxed_slice();
    let ptr = boxed.as_mut_ptr();
    unsafe {
        *out_len = boxed.len();
    }
    std::mem::forget(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn ssh_core_free_bytes(data: *mut u8, len: usize) {
    if data.is_null() {
        return;
    }
    let _ = unsafe { Vec::from_raw_parts(data, len, len) };
}

#[no_mangle]
pub extern "C" fn ssh_core_create_vault(password: *const c_char) -> *mut Vault {
    if password.is_null() {
        return std::ptr::null_mut();
    }

    let password_str = unsafe {
        match CStr::from_ptr(password).to_str() {
            Ok(s) => s,
            Err(_) => return std::ptr::null_mut(),
        }
    };

    let secret = SecretString::new(password_str.to_string());
    match Vault::create(&secret) {
        Ok(vault) => Box::into_raw(Box::new(vault)),
        Err(_) => std::ptr::null_mut(),
    }
}

#[no_mangle]
pub extern "C" fn ssh_core_unlock_vault(
    data_json: *const c_char,
    password: *const c_char,
) -> *mut Vault {
    if data_json.is_null() || password.is_null() {
        return std::ptr::null_mut();
    }

    let json_str = unsafe {
        match CStr::from_ptr(data_json).to_str() {
            Ok(s) => s,
            Err(_) => return std::ptr::null_mut(),
        }
    };

    let password_str = unsafe {
        match CStr::from_ptr(password).to_str() {
            Ok(s) => s,
            Err(_) => return std::ptr::null_mut(),
        }
    };

    let vault_data: VaultData = match serde_json::from_str(json_str) {
        Ok(d) => d,
        Err(_) => return std::ptr::null_mut(),
    };

    let secret = SecretString::new(password_str.to_string());
    match Vault::unlock(vault_data, &secret) {
        Ok(vault) => Box::into_raw(Box::new(vault)),
        Err(_) => std::ptr::null_mut(),
    }
}

#[no_mangle]
pub extern "C" fn ssh_core_free_vault(vault: *mut Vault) {
    if !vault.is_null() {
        unsafe {
            let _ = Box::from_raw(vault);
        }
    }
}
