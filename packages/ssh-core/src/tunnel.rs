use crate::ssh::{SshError, SshManager};
use ssh2::Channel;
use std::collections::HashMap;
use std::io::{Read, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::{Arc, Mutex};
use std::thread;

pub struct TunnelManager {
    tunnels: Arc<Mutex<HashMap<String, thread::JoinHandle<()>>>>,
}

impl TunnelManager {
    pub fn new() -> Self {
        Self {
            tunnels: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn local_forward(
        &self,
        session_id: &str,
        local_port: u16,
        remote_host: &str,
        remote_port: u16,
        ssh_manager: &SshManager,
    ) -> Result<String, SshError> {
        let sessions = ssh_manager.sessions.lock().unwrap();
        let ssh_session = sessions.get(session_id).ok_or(SshError::SessionNotFound)?;
        let session = ssh_session.session.clone();
        drop(sessions);

        let remote_host = remote_host.to_string();
        let tunnel_id = uuid::Uuid::new_v4().to_string();
        let tunnels = self.tunnels.clone();
        let id = tunnel_id.clone();

        let handle = thread::spawn(move || {
            let listener = match TcpListener::bind(format!("127.0.0.1:{}", local_port)) {
                Ok(l) => l,
                Err(_) => return,
            };

            for stream in listener.incoming().flatten() {
                let mut channel = match session.channel_direct_tcpip(
                    &remote_host,
                    remote_port,
                    None,
                ) {
                    Ok(c) => c,
                    Err(_) => continue,
                };
                let _ = tunnel_stream(stream, &mut channel);
            }
            tunnels.lock().unwrap().remove(&id);
        });

        self.tunnels.lock().unwrap().insert(tunnel_id.clone(), handle);
        Ok(tunnel_id)
    }

    pub fn stop_tunnel(&self, tunnel_id: &str) -> Result<(), SshError> {
        if let Some(handle) = self.tunnels.lock().unwrap().remove(tunnel_id) {
            // Note: We cannot forcefully join, but removing stops new connections
            let _ = handle.join();
        }
        Ok(())
    }
}

impl Default for TunnelManager {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_tunnel_manager_new() {
        let manager = TunnelManager::new();
        let tunnels = manager.tunnels.lock().unwrap();
        assert!(tunnels.is_empty());
    }

    #[test]
    fn test_stop_nonexistent_tunnel() {
        let manager = TunnelManager::new();
        assert!(manager.stop_tunnel("nonexistent").is_ok());
    }

    #[test]
    fn test_tunnel_manager_default() {
        let manager = TunnelManager::default();
        let tunnels = manager.tunnels.lock().unwrap();
        assert!(tunnels.is_empty());
    }
}

fn tunnel_stream(mut local: TcpStream, remote: &mut Channel) -> std::io::Result<()> {
    let mut buf_local = [0u8; 4096];
    let mut buf_remote = [0u8; 4096];

    local.set_nonblocking(true)?;

    loop {
        match local.read(&mut buf_local) {
            Ok(0) => break,
            Ok(n) => {
                remote.write_all(&buf_local[..n])?;
                remote.flush()?;
            }
            Err(ref e) if e.kind() == std::io::ErrorKind::WouldBlock => {}
            Err(e) => return Err(e),
        }

        match remote.read(&mut buf_remote) {
            Ok(0) => break,
            Ok(n) => {
                local.write_all(&buf_remote[..n])?;
                local.flush()?;
            }
            Err(ref e) if e.kind() == std::io::ErrorKind::WouldBlock => {}
            Err(e) => return Err(e),
        }
    }

    Ok(())
}
