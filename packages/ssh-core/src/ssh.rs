use secrecy::{ExposeSecret, SecretString};
use std::collections::HashMap;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::sync::{Arc, Mutex};
use ssh2::Session;

#[derive(Debug, thiserror::Error)]
pub enum SshError {
    #[error("connection failed")]
    Connect,
    #[error("authentication failed")]
    Auth,
    #[error("session not found")]
    SessionNotFound,
    #[error("io error")]
    Io(#[from] std::io::Error),
    #[error("ssh2 error")]
    Ssh2(#[from] ssh2::Error),
}

pub struct SshSession {
    pub id: String,
    pub session: Session,
    pub channel: Option<ssh2::Channel>,
}

pub struct SshManager {
    pub sessions: Arc<Mutex<HashMap<String, SshSession>>>,
}

impl SshManager {
    pub fn new() -> Self {
        Self {
            sessions: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn connect(
        &self,
        host: &str,
        port: u16,
        username: &str,
        password: &SecretString,
    ) -> Result<String, SshError> {
        let tcp = TcpStream::connect(format!("{}:{}", host, port))?;
        let mut session = Session::new()?;
        session.set_tcp_stream(tcp);
        session.handshake()?;
        session.userauth_password(username, password.expose_secret())?;
        if !session.authenticated() {
            return Err(SshError::Auth);
        }
        let id = uuid::Uuid::new_v4().to_string();
        let ssh_session = SshSession {
            id: id.clone(),
            session,
            channel: None,
        };
        self.sessions.lock().unwrap().insert(id.clone(), ssh_session);
        Ok(id)
    }

    pub fn disconnect(&self, session_id: &str) -> Result<(), SshError> {
        let mut sessions = self.sessions.lock().unwrap();
        if let Some(mut ssh_session) = sessions.remove(session_id) {
            if let Some(mut channel) = ssh_session.channel.take() {
                let _ = channel.send_eof();
                let _ = channel.wait_eof();
                let _ = channel.close();
                let _ = channel.wait_close();
            }
            // Session is dropped here, closing the connection
        }
        Ok(())
    }

    pub fn write(&self, session_id: &str, data: &[u8]) -> Result<usize, SshError> {
        let mut sessions = self.sessions.lock().unwrap();
        let ssh_session = sessions.get_mut(session_id).ok_or(SshError::SessionNotFound)?;
        
        if ssh_session.channel.is_none() {
            let channel = ssh_session.session.channel_session()?;
            ssh_session.channel = Some(channel);
            if let Some(ref mut ch) = ssh_session.channel {
                ch.request_pty("xterm-256color", None, None)?;
                ch.shell()?;
            }
        }
        
        if let Some(ref mut channel) = ssh_session.channel {
            Ok(channel.write(data)?)
        } else {
            Err(SshError::Connect)
        }
    }

    pub fn read(&self, session_id: &str, buffer: &mut [u8]) -> Result<usize, SshError> {
        let mut sessions = self.sessions.lock().unwrap();
        let ssh_session = sessions.get_mut(session_id).ok_or(SshError::SessionNotFound)?;
        
        if let Some(ref mut channel) = ssh_session.channel {
            Ok(channel.read(buffer)?)
        } else {
            Err(SshError::Connect)
        }
    }

    pub fn resize(&self, session_id: &str, cols: u32, rows: u32) -> Result<(), SshError> {
        let mut sessions = self.sessions.lock().unwrap();
        let ssh_session = sessions.get_mut(session_id).ok_or(SshError::SessionNotFound)?;
        
        if let Some(ref mut channel) = ssh_session.channel {
            channel.request_pty_size(cols, rows, None, None)?;
            Ok(())
        } else {
            Err(SshError::Connect)
        }
    }
}

impl Default for SshManager {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ssh_manager_new() {
        let manager = SshManager::new();
        let sessions = manager.sessions.lock().unwrap();
        assert!(sessions.is_empty());
    }

    #[test]
    fn test_disconnect_nonexistent() {
        let manager = SshManager::new();
        // Should not panic
        let _ = manager.disconnect("nonexistent-id");
    }
}
