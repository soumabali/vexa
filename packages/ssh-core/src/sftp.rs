use crate::ssh::{SshError, SshManager};
use ssh2::Sftp;
use std::collections::HashMap;
use std::io::{Read, Write};
use std::sync::{Arc, Mutex};

pub struct SftpManager {
    sftp_sessions: Arc<Mutex<HashMap<String, Sftp>>>,
}

impl SftpManager {
    pub fn new() -> Self {
        Self {
            sftp_sessions: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn open(&self, session_id: &str, ssh_manager: &SshManager) -> Result<(), SshError> {
        let sessions = ssh_manager.sessions.lock().unwrap();
        let ssh_session = sessions.get(session_id).ok_or(SshError::SessionNotFound)?;
        let sftp = ssh_session.session.sftp()?;
        drop(sessions);
        self.sftp_sessions.lock().unwrap().insert(session_id.to_string(), sftp);
        Ok(())
    }

    pub fn close(&self, session_id: &str) -> Result<(), SshError> {
        self.sftp_sessions.lock().unwrap().remove(session_id);
        Ok(())
    }

    pub fn read_file(&self, session_id: &str, path: &str) -> Result<Vec<u8>, SshError> {
        let mut sftp_sessions = self.sftp_sessions.lock().unwrap();
        let sftp = sftp_sessions.get(session_id).ok_or(SshError::SessionNotFound)?;
        let mut file = sftp.open(std::path::Path::new(path))?;
        let mut contents = Vec::new();
        file.read_to_end(&mut contents)?;
        Ok(contents)
    }

    pub fn write_file(&self,
        session_id: &str,
        path: &str,
        data: &[u8],
    ) -> Result<(), SshError> {
        let mut sftp_sessions = self.sftp_sessions.lock().unwrap();
        let sftp = sftp_sessions.get(session_id).ok_or(SshError::SessionNotFound)?;
        let mut file = sftp.create(std::path::Path::new(path))?;
        file.write_all(data)?;
        Ok(())
    }

    pub fn list_dir(&self, session_id: &str, path: &str) -> Result<Vec<String>, SshError> {
        let sftp_sessions = self.sftp_sessions.lock().unwrap();
        let sftp = sftp_sessions.get(session_id).ok_or(SshError::SessionNotFound)?;
        let entries = sftp.readdir(std::path::Path::new(path))?;
        let names = entries
            .into_iter()
            .map(|(path, _stat)| path.to_string_lossy().to_string())
            .collect();
        Ok(names)
    }
}

impl Default for SftpManager {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sftp_manager_new() {
        let manager = SftpManager::new();
        let sessions = manager.sftp_sessions.lock().unwrap();
        assert!(sessions.is_empty());
    }

    #[test]
    fn test_close_nonexistent_session() {
        let manager = SftpManager::new();
        assert!(manager.close("nonexistent").is_ok());
    }

    #[test]
    fn test_read_file_nonexistent_session() {
        let manager = SftpManager::new();
        let result = manager.read_file("nonexistent", "/path");
        assert!(matches!(result, Err(SshError::SessionNotFound)));
    }

    #[test]
    fn test_write_file_nonexistent_session() {
        let manager = SftpManager::new();
        let result = manager.write_file("nonexistent", "/path", b"data");
        assert!(matches!(result, Err(SshError::SessionNotFound)));
    }

    #[test]
    fn test_list_dir_nonexistent_session() {
        let manager = SftpManager::new();
        let result = manager.list_dir("nonexistent", "/");
        assert!(matches!(result, Err(SshError::SessionNotFound)));
    }

    #[test]
    fn test_sftp_manager_default() {
        let manager = SftpManager::default();
        let sessions = manager.sftp_sessions.lock().unwrap();
        assert!(sessions.is_empty());
    }
}
