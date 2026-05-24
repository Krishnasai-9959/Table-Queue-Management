/* ============================================================
   MISE — script.js
   Modular: Auth · API · Toast · Btn · Guard · WS · Utils
   ============================================================ */
'use strict';

const BASE = 'http://localhost:8080';

/* ── Auth ── */
const Auth = {
  token:  () => localStorage.getItem('mise_token'),
  uid:    () => localStorage.getItem('mise_uid'),
  name:   () => localStorage.getItem('mise_name') || '',
  role:   () => localStorage.getItem('mise_role') || '',
  save({ token, user_id, role, name }) {
    if (token)   localStorage.setItem('mise_token', token);
    if (user_id) localStorage.setItem('mise_uid', user_id);
    if (role)    localStorage.setItem('mise_role', role);
    if (name)    localStorage.setItem('mise_name', name);
  },
  clear() { ['mise_token','mise_uid','mise_role','mise_name'].forEach(k => localStorage.removeItem(k)); },
  ok: () => !!localStorage.getItem('mise_token'),
};

/* ── API ── */
const API = {
  async _req(path, opts = {}) {
    const tok = Auth.token();
    const res = await fetch(`${BASE}${path}`, {
      headers: { 'Content-Type': 'application/json', ...(tok ? { Authorization: `Bearer ${tok}` } : {}), ...opts.headers },
      ...opts,
    });
    if (!res.ok) { const t = await res.text().catch(() => `HTTP ${res.status}`); throw new Error(t || `HTTP ${res.status}`); }
    const txt = await res.text();
    return txt ? JSON.parse(txt) : {};
  },
  signup:  d  => API._req('/signup',            { method: 'POST', body: JSON.stringify(d) }),
  login:   d  => API._req('/login',             { method: 'POST', body: JSON.stringify(d) }),
  logout:  () => API._req('/logout'),
  tables:  () => API._req('/tables'),
  queue:   () => API._req('/queue'),
  setAvail: v => API._req('/waiter/status',     { method: 'POST', body: JSON.stringify({ available: v }) }),
  setTableStatus: d => API._req('/table/status', { method: 'POST', body: JSON.stringify(d) }),
  waiters: () => API._req('/admin/waiters'),
  stats:   () => API._req('/admin/stats'),
  assign:  d  => API._req('/admin/assign-table',{ method: 'POST', body: JSON.stringify(d) }),
  delWaiter: id => API._req(`/admin/delete-waiter?emp_id=${id}`, { method: 'DELETE' }),
};

/* ── Toast ── */
const Toast = {
  _el: null,
  _init() {
    if (!this._el) {
      this._el = document.createElement('div');
      this._el.id = 'toasts';
      document.body.appendChild(this._el);
    }
  },
  show(msg, type = 'info', dur = 3500) {
    this._init();
    const t = document.createElement('div');
    t.className = `toast ${type}`;
    t.innerHTML = `<div class="t-dot"></div><span class="t-msg">${msg}</span><span class="t-x">✕</span>`;
    t.querySelector('.t-x').onclick = () => this._hide(t);
    this._el.appendChild(t);
    setTimeout(() => this._hide(t), dur);
  },
  _hide(t) { t.classList.add('out'); t.addEventListener('animationend', () => t.remove(), { once: true }); },
  success: (m, d) => Toast.show(m, 'success', d),
  error:   (m, d) => Toast.show(m, 'error', d),
  info:    (m, d) => Toast.show(m, 'info', d),
};

/* ── Btn loading ── */
const Btn = {
  load: b  => { b.classList.add('loading'); b.disabled = true; },
  reset: b => { b.classList.remove('loading'); b.disabled = false; },
};

/* ── Guard ── */
const Guard = {
  auth()  { if (!Auth.ok()) { location.href = 'index.html'; return false; } return true; },
  guest() { if (Auth.ok())  { location.href = 'dashboard.html'; return false; } return true; },
};

/* ── WS ── */
const WS = {
  conn: null, cbs: [],
  connect() {
    try {
      const url = `ws://${BASE.replace(/^https?:\/\//, '')}/ws`;
      this.conn = new WebSocket(url);
      this.conn.onopen    = () => WS._badge(true);
      this.conn.onmessage = e  => { WS._log(e.data); WS.cbs.forEach(f => f(e.data)); };
      this.conn.onclose   = () => { WS._badge(false); setTimeout(() => WS.connect(), 5000); };
      this.conn.onerror   = () => WS._badge(false);
    } catch (_) { WS._badge(false); }
  },
  on: f => WS.cbs.push(f),
  _badge(live) {
    const p = document.querySelector('.ws-pill');
    const l = document.querySelector('.ws-lbl');
    if (!p) return;
    p.className = `ws-pill${live ? ' live' : ''}`;
    if (l) l.textContent = live ? 'Live' : 'Off';
  },
  _log(msg) {
    const fb = document.getElementById('feed-body');
    if (!fb) return;
    const now = new Date();
    const ts = `${String(now.getHours()).padStart(2,'0')}:${String(now.getMinutes()).padStart(2,'0')}:${String(now.getSeconds()).padStart(2,'0')}`;
    const el = document.createElement('div');
    el.className = 'feed-evt';
    el.innerHTML = `<span class="feed-dot"></span><span class="feed-time">${ts}</span><span class="feed-msg">${_esc(msg)}</span>`;
    fb.insertBefore(el, fb.firstChild);
    while (fb.children.length > 40) fb.removeChild(fb.lastChild);
    // Remove empty state if present
    fb.querySelector('.empty') && fb.querySelector('.empty').remove();
  },
};

/* ── Utils ── */
function _esc(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function initials(n) {
  return (n || '?').split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2);
}

window.Mise = { Auth, API, Toast, Btn, Guard, WS, initials, _esc };