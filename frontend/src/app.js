const API_BASE = 'http://localhost:8080/api/v1';

let accessToken = localStorage.getItem('access_token');
let refreshToken = localStorage.getItem('refresh_token');

function switchTab(tab) {
  document.getElementById('login-form').classList.toggle('hidden', tab !== 'login');
  document.getElementById('register-form').classList.toggle('hidden', tab !== 'register');
  document.getElementById('tab-login').classList.toggle('tab-active', tab === 'login');
  document.getElementById('tab-register').classList.toggle('tab-active', tab === 'register');
  clearMessage();
}

function showMessage(msg, isError = false) {
  const el = document.getElementById('message');
  el.textContent = msg;
  el.className = `mt-4 text-sm text-center ${isError ? 'text-red-600' : 'text-green-600'}`;
  el.classList.remove('hidden');
}

function clearMessage() {
  document.getElementById('message').classList.add('hidden');
}

async function apiFetch(path, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) };
  if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;
  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const data = await res.json().catch(() => ({}));
  return { ok: res.ok, status: res.status, data };
}

async function handleLogin(e) {
  e.preventDefault();
  const email = document.getElementById('login-email').value;
  const password = document.getElementById('login-password').value;
  const { ok, data } = await apiFetch('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
  if (!ok) { showMessage(data.error || 'Login failed', true); return; }
  saveTokens(data);
  showDashboard(data.user);
}

async function handleRegister(e) {
  e.preventDefault();
  const username = document.getElementById('reg-username').value;
  const email = document.getElementById('reg-email').value;
  const password = document.getElementById('reg-password').value;
  const { ok, data } = await apiFetch('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, email, password }),
  });
  if (!ok) { showMessage(data.error || 'Registration failed', true); return; }
  saveTokens(data);
  showDashboard(data.user);
}

async function googleLogin() {
  const { ok, data } = await apiFetch('/auth/oauth/google');
  if (ok && data.url) window.location.href = data.url;
  else showMessage('Google login unavailable — configure OAuth credentials', true);
}

function saveTokens(data) {
  accessToken = data.access_token;
  refreshToken = data.refresh_token;
  localStorage.setItem('access_token', accessToken);
  localStorage.setItem('refresh_token', refreshToken);
}

function showDashboard(user) {
  document.getElementById('auth-card').classList.add('hidden');
  document.getElementById('dashboard').classList.remove('hidden');
  document.getElementById('user-info').textContent = `${user.username} · ${user.email}`;
  document.getElementById('user-role').textContent = user.role;
  document.getElementById('token-preview').textContent = accessToken.slice(0, 60) + '...';
  if (user.role === 'admin') {
    document.getElementById('admin-panel').classList.remove('hidden');
  }
}

async function loadUsers() {
  const { ok, data } = await apiFetch('/admin/users');
  const el = document.getElementById('users-list');
  if (!ok) { el.innerHTML = `<p class="text-red-500">${data.error}</p>`; return; }
  el.innerHTML = data.users.map(u => `
    <div class="flex justify-between items-center bg-gray-50 rounded-lg p-2">
      <span class="font-medium">${u.username}</span>
      <span class="text-xs bg-indigo-100 text-indigo-700 px-2 py-0.5 rounded-full">${u.role}</span>
    </div>
  `).join('');
}

async function logout() {
  if (refreshToken) {
    await apiFetch('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
  }
  accessToken = null;
  refreshToken = null;
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  document.getElementById('dashboard').classList.add('hidden');
  document.getElementById('auth-card').classList.remove('hidden');
  clearMessage();
}

// Auto-restore session on load
(async () => {
  if (!accessToken) return;
  const { ok, data } = await apiFetch('/me');
  if (ok) showDashboard(data);
  else {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  }
})();
