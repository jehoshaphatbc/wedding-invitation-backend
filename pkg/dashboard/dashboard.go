package dashboard

const HTML = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Wedding Invitation Backend - Live API Monitor</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-primary: #0b0f19;
      --bg-card: rgba(22, 30, 49, 0.7);
      --bg-card-hover: rgba(30, 41, 67, 0.8);
      --border-color: rgba(255, 255, 255, 0.08);
      --border-accent: rgba(99, 102, 241, 0.3);
      --text-main: #f8fafc;
      --text-muted: #94a3b8;
      --accent-indigo: #6366f1;
      --accent-purple: #a855f7;
      --accent-pink: #ec4899;
      --accent-emerald: #10b981;
      --accent-amber: #f59e0b;
      --accent-rose: #f43f5e;
      --glass-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    body {
      font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, sans-serif;
      background-color: var(--bg-primary);
      color: var(--text-main);
      min-height: 100vh;
      background-image: 
        radial-gradient(circle at 15% 15%, rgba(99, 102, 241, 0.15) 0%, transparent 40%),
        radial-gradient(circle at 85% 85%, rgba(236, 72, 153, 0.12) 0%, transparent 40%),
        radial-gradient(circle at 50% 50%, rgba(16, 185, 129, 0.08) 0%, transparent 50%);
      background-attachment: fixed;
      padding: 2.5rem 1.5rem;
      line-height: 1.6;
    }

    .container {
      max-width: 1200px;
      margin: 0 auto;
    }

    header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 2.5rem;
      flex-wrap: wrap;
      gap: 1.5rem;
      padding-bottom: 1.5rem;
      border-bottom: 1px solid var(--border-color);
    }

    .brand {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .brand-icon {
      width: 48px;
      height: 48px;
      background: linear-gradient(135deg, var(--accent-indigo), var(--accent-purple));
      border-radius: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
      font-size: 1.5rem;
    }

    .brand-title h1 {
      font-size: 1.5rem;
      font-weight: 800;
      background: linear-gradient(to right, #ffffff, #cbd5e1);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      letter-spacing: -0.02em;
    }

    .brand-title p {
      font-size: 0.875rem;
      color: var(--text-muted);
    }

    .status-badge-main {
      display: inline-flex;
      align-items: center;
      gap: 0.75rem;
      background: rgba(16, 185, 129, 0.1);
      border: 1px solid rgba(16, 185, 129, 0.3);
      padding: 0.6rem 1.25rem;
      border-radius: 9999px;
      backdrop-filter: blur(10px);
    }

    .status-ping {
      position: relative;
      display: flex;
      height: 10px;
      width: 10px;
    }

    .status-ping span:first-child {
      animation: ping 1.5s cubic-bezier(0, 0, 0.2, 1) infinite;
      position: absolute;
      display: inline-flex;
      height: 100%;
      width: 100%;
      border-radius: 50%;
      background-color: var(--accent-emerald);
      opacity: 0.75;
    }

    .status-ping span:last-child {
      position: relative;
      display: inline-flex;
      border-radius: 50%;
      height: 10px;
      width: 10px;
      background-color: var(--accent-emerald);
    }

    @keyframes ping {
      75%, 100% {
        transform: scale(2.2);
        opacity: 0;
      }
    }

    .status-text {
      font-weight: 600;
      font-size: 0.9rem;
      color: var(--accent-emerald);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .overview-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
      gap: 1.25rem;
      margin-bottom: 2.5rem;
    }

    .card {
      background: var(--bg-card);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border: 1px solid var(--border-color);
      border-radius: 16px;
      padding: 1.5rem;
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      box-shadow: var(--glass-shadow);
    }

    .card:hover {
      background: var(--bg-card-hover);
      border-color: var(--border-accent);
      transform: translateY(-2px);
    }

    .card-label {
      font-size: 0.8rem;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--text-muted);
      margin-bottom: 0.5rem;
      font-weight: 600;
    }

    .card-value {
      font-size: 1.4rem;
      font-weight: 700;
      color: #fff;
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .card-subtext {
      font-size: 0.825rem;
      color: var(--accent-emerald);
      margin-top: 0.35rem;
    }

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
    }

    .section-title {
      font-size: 1.25rem;
      font-weight: 700;
      color: #fff;
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .btn-run-all {
      background: linear-gradient(135deg, var(--accent-indigo), var(--accent-purple));
      color: #fff;
      border: none;
      padding: 0.75rem 1.5rem;
      border-radius: 12px;
      font-weight: 600;
      font-size: 0.9rem;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 0.5rem;
      transition: all 0.2s ease;
      box-shadow: 0 4px 15px rgba(99, 102, 241, 0.4);
    }

    .btn-run-all:hover {
      transform: scale(1.02);
      box-shadow: 0 6px 20px rgba(99, 102, 241, 0.6);
    }

    .btn-run-all:active {
      transform: scale(0.98);
    }

    .test-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
      gap: 1.5rem;
    }

    .test-card {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: 16px;
      padding: 1.5rem;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      transition: all 0.3s ease;
    }

    .test-card-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1rem;
    }

    .method-badge {
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.75rem;
      font-weight: 700;
      padding: 0.25rem 0.6rem;
      border-radius: 6px;
      text-transform: uppercase;
    }

    .method-get {
      background: rgba(16, 185, 129, 0.15);
      color: var(--accent-emerald);
      border: 1px solid rgba(16, 185, 129, 0.3);
    }

    .method-post {
      background: rgba(99, 102, 241, 0.15);
      color: var(--accent-indigo);
      border: 1px solid rgba(99, 102, 241, 0.3);
    }

    .test-endpoint {
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.95rem;
      font-weight: 600;
      color: #fff;
      margin-top: 0.35rem;
    }

    .test-desc {
      font-size: 0.85rem;
      color: var(--text-muted);
      margin-bottom: 1rem;
    }

    .test-status-badge {
      font-size: 0.75rem;
      font-weight: 700;
      padding: 0.3rem 0.75rem;
      border-radius: 9999px;
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
    }

    .badge-pending {
      background: rgba(148, 163, 184, 0.1);
      color: var(--text-muted);
      border: 1px solid rgba(148, 163, 184, 0.2);
    }

    .badge-success {
      background: rgba(16, 185, 129, 0.15);
      color: var(--accent-emerald);
      border: 1px solid rgba(16, 185, 129, 0.3);
    }

    .badge-error {
      background: rgba(244, 63, 94, 0.15);
      color: var(--accent-rose);
      border: 1px solid rgba(244, 63, 94, 0.3);
    }

    .response-box {
      background: rgba(8, 11, 19, 0.8);
      border: 1px solid rgba(255, 255, 255, 0.05);
      border-radius: 10px;
      padding: 0.85rem;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.8rem;
      color: #cbd5e1;
      max-height: 140px;
      overflow-y: auto;
      margin-top: 0.75rem;
      white-space: pre-wrap;
      word-break: break-all;
    }

    .btn-test {
      width: 100%;
      margin-top: 1rem;
      background: rgba(255, 255, 255, 0.05);
      color: #fff;
      border: 1px solid var(--border-color);
      padding: 0.6rem;
      border-radius: 10px;
      font-size: 0.85rem;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.2s ease;
    }

    .btn-test:hover {
      background: rgba(255, 255, 255, 0.1);
      border-color: rgba(255, 255, 255, 0.2);
    }

    footer {
      margin-top: 3rem;
      text-align: center;
      font-size: 0.85rem;
      color: var(--text-muted);
      border-top: 1px solid var(--border-color);
      padding-top: 1.5rem;
    }
  </style>
</head>
<body>
  <div class="container">
    <header>
      <div class="brand">
        <div class="brand-icon">💍</div>
        <div class="brand-title">
          <h1>Wedding Invitation Backend API</h1>
          <p>Go 1.26 + Gin Framework + PostgreSQL + GORM</p>
        </div>
      </div>
      <div class="status-badge-main" id="mainStatusBadge">
        <div class="status-ping">
          <span></span>
          <span></span>
        </div>
        <span class="status-text">SYSTEM ONLINE &amp; HEALTHY</span>
      </div>
    </header>

    <div class="overview-grid">
      <div class="card">
        <div class="card-label">Server Status</div>
        <div class="card-value" style="color: var(--accent-emerald);">HTTP 200 OK</div>
        <div class="card-subtext">Port 8080 Active</div>
      </div>
      <div class="card">
        <div class="card-label">Database Connection</div>
        <div class="card-value" style="color: var(--accent-indigo);">PostgreSQL</div>
        <div class="card-subtext">GORM Auto-Migrated</div>
      </div>
      <div class="card">
        <div class="card-label">Authentication</div>
        <div class="card-value" style="color: var(--accent-purple);">JWT Bearer</div>
        <div class="card-subtext">RBAC Roles &amp; Permissions Enabled</div>
      </div>
      <div class="card">
        <div class="card-label">Super Admin Account</div>
        <div class="card-value" style="font-size: 1rem; color: var(--accent-amber);">admin@wedding.com</div>
        <div class="card-subtext">Seeded &amp; Verified</div>
      </div>
    </div>

    <div class="section-header">
      <div class="section-title">
        <span>🧪 Live API Verification Console</span>
      </div>
      <button class="btn-run-all" onclick="runAllTests()">⚡ Run All Endpoint Tests</button>
    </div>

    <div class="test-grid">
      <div class="test-card">
        <div>
          <div class="test-card-header">
            <div>
              <span class="method-badge method-get">GET</span>
              <div class="test-endpoint">/health</div>
            </div>
            <span class="test-status-badge badge-pending" id="status-health">Ready</span>
          </div>
          <div class="test-desc">Verifikasi kesehatan server backend dan ketersediaan service API.</div>
        </div>
        <div>
          <div class="response-box" id="res-health">// Click run to test...</div>
          <button class="btn-test" onclick="testHealth()">Test Health Check</button>
        </div>
      </div>

      <div class="test-card">
        <div>
          <div class="test-card-header">
            <div>
              <span class="method-badge method-post">POST</span>
              <div class="test-endpoint">/api/v1/auth/login</div>
            </div>
            <span class="test-status-badge badge-pending" id="status-login">Ready</span>
          </div>
          <div class="test-desc">Autentikasi Super Admin dan penerbitan token JWT (Access &amp; Refresh token).</div>
        </div>
        <div>
          <div class="response-box" id="res-login">// Click run to test...</div>
          <button class="btn-test" onclick="testLogin()">Test Login</button>
        </div>
      </div>

      <div class="test-card">
        <div>
          <div class="test-card-header">
            <div>
              <span class="method-badge method-get">GET</span>
              <div class="test-endpoint">/api/v1/me</div>
            </div>
            <span class="test-status-badge badge-pending" id="status-me">Ready</span>
          </div>
          <div class="test-desc">Mengambil profil akun user yang sedang login dengan Bearer JWT token.</div>
        </div>
        <div>
          <div class="response-box" id="res-me">// Requires login token...</div>
          <button class="btn-test" onclick="testProfile()">Test Profile</button>
        </div>
      </div>

      <div class="test-card">
        <div>
          <div class="test-card-header">
            <div>
              <span class="method-badge method-get">GET</span>
              <div class="test-endpoint">/api/v1/admin/stats</div>
            </div>
            <span class="test-status-badge badge-pending" id="status-stats">Ready</span>
          </div>
          <div class="test-desc">Dashboard statistik pengguna dan sistem (Requires Admin Role).</div>
        </div>
        <div>
          <div class="response-box" id="res-stats">// Requires login token...</div>
          <button class="btn-test" onclick="testStats()">Test Admin Stats</button>
        </div>
      </div>

      <div class="test-card">
        <div>
          <div class="test-card-header">
            <div>
              <span class="method-badge method-get">GET</span>
              <div class="test-endpoint">/api/v1/admin/roles</div>
            </div>
            <span class="test-status-badge badge-pending" id="status-roles">Ready</span>
          </div>
          <div class="test-desc">Mendapatkan daftar role RBAC dari database PostgreSQL (Super Admin, Admin, Customer).</div>
        </div>
        <div>
          <div class="response-box" id="res-roles">// Requires login token...</div>
          <button class="btn-test" onclick="testRoles()">Test Admin Roles</button>
        </div>
      </div>

      <div class="test-card">
        <div>
          <div class="test-card-header">
            <div>
              <span class="method-badge method-get">GET</span>
              <div class="test-endpoint">/api/v1/admin/permissions</div>
            </div>
            <span class="test-status-badge badge-pending" id="status-permissions">Ready</span>
          </div>
          <div class="test-desc">Mendapatkan semua hak akses (permissions) terdaftar dalam sistem.</div>
        </div>
        <div>
          <div class="response-box" id="res-permissions">// Requires login token...</div>
          <button class="btn-test" onclick="testPermissions()">Test Permissions</button>
        </div>
      </div>
    </div>

    <footer>
      Wedding Invitation Backend API &bull; Running on <strong>http://localhost:8080</strong> &bull; Operational Status: <strong>NORMAL</strong>
    </footer>
  </div>

  <script>
    var jwtToken = '';

    function updateBadge(id, status, text) {
      var el = document.getElementById(id);
      el.className = 'test-status-badge badge-' + status;
      el.innerText = text;
    }

    function testHealth() {
      updateBadge('status-health', 'pending', 'Testing...');
      var startTime = performance.now();
      return fetch('/health')
        .then(function(res) {
          return res.json().then(function(data) {
            var duration = Math.round(performance.now() - startTime);
            document.getElementById('res-health').innerText = '[' + res.status + ' OK] (' + duration + 'ms)\n' + JSON.stringify(data, null, 2);
            updateBadge('status-health', 'success', 'SUCCESS (' + duration + 'ms)');
            return true;
          });
        })
        .catch(function(err) {
          document.getElementById('res-health').innerText = 'Error: ' + err.message;
          updateBadge('status-health', 'error', 'FAILED');
          return false;
        });
    }

    function testLogin() {
      updateBadge('status-login', 'pending', 'Testing...');
      var startTime = performance.now();
      return fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'admin@wedding.com',
          password: 'SuperAdmin123!'
        })
      })
      .then(function(res) {
        return res.json().then(function(data) {
          var duration = Math.round(performance.now() - startTime);
          if (data.data && data.data.access_token) {
            jwtToken = data.data.access_token;
          }
          document.getElementById('res-login').innerText = '[' + res.status + ' ' + res.statusText + '] (' + duration + 'ms)\n' + JSON.stringify(data, null, 2);
          if (res.ok) {
            updateBadge('status-login', 'success', 'SUCCESS (' + duration + 'ms)');
            return true;
          } else {
            updateBadge('status-login', 'error', 'HTTP ' + res.status);
            return false;
          }
        });
      })
      .catch(function(err) {
        document.getElementById('res-login').innerText = 'Error: ' + err.message;
        updateBadge('status-login', 'error', 'FAILED');
        return false;
      });
    }

    function testProfile() {
      var p = jwtToken ? Promise.resolve() : testLogin();
      return p.then(function() {
        updateBadge('status-me', 'pending', 'Testing...');
        var startTime = performance.now();
        return fetch('/api/v1/me', {
          headers: { 'Authorization': 'Bearer ' + jwtToken }
        })
        .then(function(res) {
          return res.json().then(function(data) {
            var duration = Math.round(performance.now() - startTime);
            document.getElementById('res-me').innerText = '[' + res.status + ' OK] (' + duration + 'ms)\n' + JSON.stringify(data, null, 2);
            if (res.ok) {
              updateBadge('status-me', 'success', 'SUCCESS (' + duration + 'ms)');
              return true;
            } else {
              updateBadge('status-me', 'error', 'HTTP ' + res.status);
              return false;
            }
          });
        })
        .catch(function(err) {
          document.getElementById('res-me').innerText = 'Error: ' + err.message;
          updateBadge('status-me', 'error', 'FAILED');
          return false;
        });
      });
    }

    function testStats() {
      var p = jwtToken ? Promise.resolve() : testLogin();
      return p.then(function() {
        updateBadge('status-stats', 'pending', 'Testing...');
        var startTime = performance.now();
        return fetch('/api/v1/admin/stats', {
          headers: { 'Authorization': 'Bearer ' + jwtToken }
        })
        .then(function(res) {
          return res.json().then(function(data) {
            var duration = Math.round(performance.now() - startTime);
            document.getElementById('res-stats').innerText = '[' + res.status + ' OK] (' + duration + 'ms)\n' + JSON.stringify(data, null, 2);
            if (res.ok) {
              updateBadge('status-stats', 'success', 'SUCCESS (' + duration + 'ms)');
              return true;
            } else {
              updateBadge('status-stats', 'error', 'HTTP ' + res.status);
              return false;
            }
          });
        })
        .catch(function(err) {
          document.getElementById('res-stats').innerText = 'Error: ' + err.message;
          updateBadge('status-stats', 'error', 'FAILED');
          return false;
        });
      });
    }

    function testRoles() {
      var p = jwtToken ? Promise.resolve() : testLogin();
      return p.then(function() {
        updateBadge('status-roles', 'pending', 'Testing...');
        var startTime = performance.now();
        return fetch('/api/v1/admin/roles', {
          headers: { 'Authorization': 'Bearer ' + jwtToken }
        })
        .then(function(res) {
          return res.json().then(function(data) {
            var duration = Math.round(performance.now() - startTime);
            document.getElementById('res-roles').innerText = '[' + res.status + ' OK] (' + duration + 'ms)\n' + JSON.stringify(data, null, 2);
            if (res.ok) {
              updateBadge('status-roles', 'success', 'SUCCESS (' + duration + 'ms)');
              return true;
            } else {
              updateBadge('status-roles', 'error', 'HTTP ' + res.status);
              return false;
            }
          });
        })
        .catch(function(err) {
          document.getElementById('res-roles').innerText = 'Error: ' + err.message;
          updateBadge('status-roles', 'error', 'FAILED');
          return false;
        });
      });
    }

    function testPermissions() {
      var p = jwtToken ? Promise.resolve() : testLogin();
      return p.then(function() {
        updateBadge('status-permissions', 'pending', 'Testing...');
        var startTime = performance.now();
        return fetch('/api/v1/admin/permissions', {
          headers: { 'Authorization': 'Bearer ' + jwtToken }
        })
        .then(function(res) {
          return res.json().then(function(data) {
            var duration = Math.round(performance.now() - startTime);
            document.getElementById('res-permissions').innerText = '[' + res.status + ' OK] (' + duration + 'ms)\n' + JSON.stringify(data, null, 2);
            if (res.ok) {
              updateBadge('status-permissions', 'success', 'SUCCESS (' + duration + 'ms)');
              return true;
            } else {
              updateBadge('status-permissions', 'error', 'HTTP ' + res.status);
              return false;
            }
          });
        })
        .catch(function(err) {
          document.getElementById('res-permissions').innerText = 'Error: ' + err.message;
          updateBadge('status-permissions', 'error', 'FAILED');
          return false;
        });
      });
    }

    function runAllTests() {
      testHealth().then(function() {
        return testLogin();
      }).then(function() {
        return testProfile();
      }).then(function() {
        return testStats();
      }).then(function() {
        return testRoles();
      }).then(function() {
        return testPermissions();
      });
    }

    window.addEventListener('DOMContentLoaded', function() {
      runAllTests();
    });
  </script>
</body>
</html>`
