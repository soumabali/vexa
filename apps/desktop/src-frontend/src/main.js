(async () => {
    const status = document.getElementById('status');
    const version = document.getElementById('version');

    try {
        // Try Tauri IPC — works in desktop app, falls back in browser
        if (window.__TAURI__) {
            const v = await window.__TAURI__.core.invoke('get_app_version');
            version.textContent = v;
            status.textContent = '✅ Backend connected';
            status.style.background = '#064e3b';
            status.style.borderColor = '#065f46';
        } else {
            throw new Error('Not in Tauri');
        }
    } catch (e) {
        version.textContent = 'dev';
        status.textContent = '⚡ Running in browser mode (use Tauri for full features)';
        status.style.background = '#1e293b';
    }
})();
