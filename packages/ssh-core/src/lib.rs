pub mod crypto;
pub mod ffi;
pub mod sftp;
pub mod ssh;
pub mod tunnel;
pub mod vault;

pub use crypto::*;
pub use ssh::{SshError, SshManager, SshSession};
pub use vault::{Credential, Vault, VaultData, VaultError};

pub fn version() -> &'static str {
    env!("CARGO_PKG_VERSION")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_version_returns_non_empty() {
        let v = version();
        assert!(!v.is_empty());
        assert!(v.contains('.'));
    }

    #[test]
    fn test_module_exports() {
        // Verify re-exports compile
        let _manager = SshManager::new();
    }
}
