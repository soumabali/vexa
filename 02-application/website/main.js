document.addEventListener('DOMContentLoaded', () => {
  // Fetch latest version
  fetch('/api/version')
    .then(r => r.json())
    .then(data => {
      document.getElementById('latest-version').textContent = data.version;
    })
    .catch(() => {});

  // Platform detection
  const ua = navigator.userAgent;
  let os = 'linux';
  if (ua.includes('Win')) os = 'windows';
  if (ua.includes('Mac')) os = 'macos';

  document.querySelectorAll('.platform').forEach(el => {
    if (el.dataset.os === os) {
      el.classList.add('highlight');
    }
  });
});
