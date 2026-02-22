/* ═══════════════════════════════════════════════════════
   EVENTIFY — Frontend Application Logic (MPA Version)
═══════════════════════════════════════════════════════ */

'use strict';

const BACKEND_URL = 'http://localhost:8080';
const API = `${BACKEND_URL}/api`;

const state = {
  token: localStorage.getItem('token') || null,
  user: JSON.parse(localStorage.getItem('user') || 'null'),
  events: [],
  filter: 'all',
};

// ─── API ─────────────────────────────────────────────────────────────
async function apiFetch(path, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...options.headers };
  if (state.token) headers['Authorization'] = `Bearer ${state.token}`;

  const res = await fetch(`${API}${path}`, { ...options, headers });
  const data = await res.json().catch(() => ({}));

  if (!res.ok) throw new Error(data.error || `Request failed (${res.status})`);
  return data;
}

const api = {
  register: (body) => apiFetch('/auth/register', { method: 'POST', body: JSON.stringify(body) }),
  login: (body) => apiFetch('/auth/login', { method: 'POST', body: JSON.stringify(body) }),
  listEvents: () => apiFetch('/events'),
  getEvent: (id) => apiFetch(`/events/${id}`),
  createEvent: (body) => apiFetch('/events', { method: 'POST', body: JSON.stringify(body) }),
  bookEvent: (id) => apiFetch(`/events/${id}/register`, { method: 'POST' }),
  myRegistrations: () => apiFetch('/me/registrations'),
};

// ─── AUTH ────────────────────────────────────────────────────────────
function saveAuth(token, user) {
  state.token = token;
  state.user = user;
  localStorage.setItem('token', token);
  localStorage.setItem('user', JSON.stringify(user));
}

function logout() {
  state.token = null;
  state.user = null;
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  window.location.href = './index.html';
}

function isLoggedIn() { return !!state.token; }

// ─── NAVBAR RENDER ───────────────────────────────────────────────────
function renderNav() {
  const guest = document.getElementById('guestActions');
  const userAct = document.getElementById('userActions');
  const authReq = document.querySelectorAll('.auth-req');
  const orgReq = document.querySelectorAll('.org-req');

  if (isLoggedIn() && state.user) {
    if (guest) guest.style.display = 'none';
    if (userAct) userAct.style.display = '';
    authReq.forEach(el => el.style.display = '');
    orgReq.forEach(el => el.style.display = state.user.role === 'organizer' ? '' : 'none');

    if (document.getElementById('avatarInitial')) {
      document.getElementById('avatarInitial').textContent = state.user.name[0].toUpperCase();
      document.getElementById('dropdownName').textContent = state.user.name;
      document.getElementById('dropdownRole').textContent = state.user.role;
    }
  } else {
    if (guest) guest.style.display = '';
    if (userAct) userAct.style.display = 'none';
    authReq.forEach(el => el.style.display = 'none');
    orgReq.forEach(el => el.style.display = 'none');
  }
}

function toggleUserMenu() {
  document.getElementById('userDropdown')?.classList.toggle('open');
}
document.addEventListener('click', (e) => {
  const menu = document.getElementById('avatarBtn');
  const drop = document.getElementById('userDropdown');
  if (menu && !menu.contains(e.target) && drop && !drop.contains(e.target)) {
    drop.classList.remove('open');
  }
});

// ─── TOASTS ──────────────────────────────────────────────────────────
function toast(message, type = 'info', duration = 3500) {
  const icons = { success: '✅', error: '❌', info: 'ℹ️' };
  const el = document.createElement('div');
  el.className = `toast toast-${type}`;
  el.innerHTML = `<span>${icons[type]}</span><span>${message}</span>`;
  document.getElementById('toastContainer')?.appendChild(el);
  setTimeout(() => el.remove(), duration);
}

// ─── UTILS ───────────────────────────────────────────────────────────
function setButtonLoading(btn, loading, label) {
  if (!btn) return;
  if (loading) {
    btn._label = btn.innerHTML;
    btn.disabled = true;
    btn.innerHTML = `<span class="spinner"></span>`;
  } else {
    btn.disabled = false;
    btn.innerHTML = label || btn._label || 'Submit';
  }
}

function escHtml(str) {
  if (!str) return '';
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

const EVENT_EMOJIS = ['🎸', '🚀', '🎨', '🎭', '💻', '🎤', '🌟', '🏆', '🎯', '🎪', '🎓', '🌊'];
function eventEmoji(id) {
  let hash = 0;
  for (const c of id) hash = (hash * 31 + c.charCodeAt(0)) | 0;
  return EVENT_EMOJIS[Math.abs(hash) % EVENT_EMOJIS.length];
}
function eventHue(id) {
  let hash = 0;
  for (const c of id) hash = (hash * 17 + c.charCodeAt(0)) | 0;
  return Math.abs(hash) % 360;
}

// ─── PAGE INITIALIZATION ROUTER ──────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  renderNav();
  const page = document.body.dataset.page;

  if (page === 'home') initHome();
  else if (page === 'event-detail') initEventDetail();
  else if (page === 'tickets') initTickets();
});

// ─── HOME PAGES (Events) ─────────────────────────────────────────────
async function initHome() {
  const grid = document.getElementById('eventsGrid');
  try {
    const data = await api.listEvents();
    state.events = data.events || [];
    renderEvents();
  } catch (err) {
    grid.innerHTML = `<p style="color:var(--text-muted);padding:2rem">${err.message}</p>`;
  }
}

function setFilter(f, btn) {
  state.filter = f;
  document.querySelectorAll('.filter-tab').forEach(t => t.classList.remove('active'));
  btn.classList.add('active');
  renderEvents();
}

function filterEvents() { renderEvents(); }

function renderEvents() {
  const grid = document.getElementById('eventsGrid');
  const q = (document.getElementById('searchInput')?.value || '').toLowerCase();
  let evts = state.events;

  if (q) evts = evts.filter(e => e.title.toLowerCase().includes(q) || (e.description || '').toLowerCase().includes(q));
  if (state.filter === 'available') evts = evts.filter(e => e.available_seats > 0);
  if (state.filter === 'full') evts = evts.filter(e => e.available_seats === 0);

  if (!evts.length) {
    grid.innerHTML = `<div class="empty-state" style="grid-column:1/-1">
      <div class="empty-icon">🔍</div><h3>No events found</h3>
      <p>Try a different search or check back later</p>
    </div>`;
    return;
  }

  grid.innerHTML = evts.map(ev => {
    const seats = ev.available_seats ?? (ev.capacity - ev.registered);
    const badgeCls = seats === 0 ? 'badge-full' : seats <= 5 ? 'badge-low' : 'badge-available';
    const badgeTxt = seats === 0 ? '🔴 Sold Out' : seats <= 5 ? `⚡ ${seats} left` : `✅ ${seats} seats`;
    const dateStr = new Date(ev.event_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' });

    return `
    <div class="event-card" onclick="window.location.href='./event.html?id=${ev.id}'">
      <div class="event-card-banner">
        <div class="event-card-banner-inner" style="--hue:${eventHue(ev.id)}">${eventEmoji(ev.id)}</div>
        <div class="capacity-badge ${badgeCls}">${badgeTxt}</div>
      </div>
      <div class="event-card-body">
        <div class="event-date">📅 ${dateStr}</div>
        <div class="event-title">${escHtml(ev.title)}</div>
        <div class="event-desc">${escHtml(ev.description || 'No description provided.')}</div>
        <div class="event-meta">
          <div class="event-seats"><strong>${seats}</strong> / ${ev.capacity} seats available</div>
          <div class="event-organizer">by ${escHtml(ev.organizer?.name || 'Organizer')}</div>
        </div>
      </div>
    </div>`;
  }).join('');
}

// ─── EVENT DETAIL PAGE ───────────────────────────────────────────────
let currentEventId = null;

async function initEventDetail() {
  const urlParams = new URLSearchParams(window.location.search);
  currentEventId = urlParams.get('id');
  if (!currentEventId) { window.location.href = './index.html'; return; }

  const container = document.getElementById('eventDetailContent');
  try {
    const ev = await api.getEvent(currentEventId);
    const seats = ev.available_seats ?? (ev.capacity - ev.registered);
    const pct = Math.round((ev.registered / ev.capacity) * 100);
    const dateStr = new Date(ev.event_date).toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });

    const bookBtn = isLoggedIn()
      ? `<button class="btn btn-success btn-lg w-full" id="bookBtn" onclick="openBookingModal('${escHtml(ev.title)}')" ${seats === 0 ? 'disabled' : ''}>${seats === 0 ? 'Sold Out 🔴' : 'Book This Event 🎟️'}</button>`
      : `<button class="btn btn-primary btn-lg w-full" onclick="window.location.href='./login.html'">Log in to Book</button>`;

    container.innerHTML = `
    <div class="event-detail-hero">
      <div class="event-detail-hero-inner" style="--dHue:${eventHue(ev.id)}">${eventEmoji(ev.id)}</div>
    </div>
    <div class="event-detail-grid">
      <div class="event-detail-info">
        <h1>${escHtml(ev.title)}</h1>
        <div class="detail-meta-row">
          <div class="detail-meta-item">📅 ${dateStr}</div>
          <div class="detail-meta-item">👥 ${ev.capacity} capacity</div>
          <div class="detail-meta-item">🎪 by ${escHtml(ev.organizer?.name || 'Organizer')}</div>
        </div>
        <div class="event-detail-desc">${escHtml(ev.description || 'No description provided.')}</div>
      </div>
      <div class="event-detail-side">
        <div class="price-label">Entry</div>
        <div class="price-tag">FREE</div>
        <div class="seat-progress">
          <div class="seat-bar-track"><div class="seat-bar-fill" style="width:${pct}%"></div></div>
          <div class="seat-counts"><span>${ev.registered} registered</span><span>${seats} remaining</span></div>
        </div>
        ${bookBtn}
      </div>
    </div>`;
  } catch (err) {
    container.innerHTML = `<p style="color:var(--text-muted)">${err.message}</p>`;
  }
}

function openBookingModal(title) {
  document.getElementById('bookingModalBody').innerHTML = `You're about to book <strong>${title}</strong>.<br/>This cannot be undone.`;
  document.getElementById('bookingModal').classList.add('open');
}
function closeModal(id) { document.getElementById(id).classList.remove('open'); }
function closeModalIfBg(e, id) { if (e.target.id === id) closeModal(id); }

async function confirmBook() {
  const btn = document.getElementById('confirmBookBtn');
  setButtonLoading(btn, true);
  try {
    await api.bookEvent(currentEventId);
    closeModal('bookingModal');
    alert("🎟️ You're booked! Seat reserved successfully.");
    window.location.reload();
  } catch (err) {
    closeModal('bookingModal');
    toast(err.message, 'error');
  } finally {
    setButtonLoading(btn, false, 'Book Now 🎟️');
  }
}

// ─── TICKETS PAGE ────────────────────────────────────────────────────
async function initTickets() {
  if (!isLoggedIn()) { window.location.href = './login.html'; return; }
  const list = document.getElementById('myTicketsList');

  try {
    const data = await api.myRegistrations();
    const regs = data.registrations || [];

    if (!regs.length) {
      list.innerHTML = `<div class="empty-state">
        <div class="empty-icon">🎫</div><h3>No tickets yet</h3>
        <p>Browse events and book your first experience</p>
        <button class="btn btn-primary" onclick="window.location.href='./index.html'">Browse Events</button>
      </div>`;
      return;
    }

    list.innerHTML = regs.map(reg => {
      const ev = reg.event || {};
      const dateStr = ev.event_date ? new Date(ev.event_date).toLocaleDateString() : '—';
      return `
      <div class="ticket-card" onclick="window.location.href='./event.html?id=${reg.event_id}'">
        <div class="ticket-icon">${eventEmoji(reg.event_id)}</div>
        <div class="ticket-info">
          <div class="ticket-title">${escHtml(ev.title || 'Event')}</div>
          <div class="ticket-date">📅 ${dateStr} · Booking ID: ${reg.id.slice(0, 8)}…</div>
        </div>
        <div class="ticket-status"><span class="status-confirmed">${reg.status}</span></div>
      </div>`;
    }).join('');
  } catch (err) {
    list.innerHTML = `<p style="color:var(--text-muted)">${err.message}</p>`;
  }
}

// ─── FORMS ───────────────────────────────────────────────────────────
async function submitLogin(e) {
  e.preventDefault();
  const btn = document.getElementById('loginBtn');
  setButtonLoading(btn, true);
  try {
    const data = await api.login({
      email: document.getElementById('loginEmail').value,
      password: document.getElementById('loginPassword').value,
    });
    saveAuth(data.token, data.user);
    window.location.href = './index.html';
  } catch (err) { toast(err.message, 'error'); }
  finally { setButtonLoading(btn, false, 'Log in'); }
}

async function submitRegister(e) {
  e.preventDefault();
  const btn = document.getElementById('registerBtn');
  const role = document.querySelector('input[name="role"]:checked').value;
  setButtonLoading(btn, true);
  try {
    const data = await api.register({
      name: document.getElementById('regName').value,
      email: document.getElementById('regEmail').value,
      password: document.getElementById('regPassword').value, role,
    });
    saveAuth(data.token, data.user);
    window.location.href = './index.html';
  } catch (err) { toast(err.message, 'error'); }
  finally { setButtonLoading(btn, false, 'Create account'); }
}

async function submitCreateEvent(e) {
  e.preventDefault();
  const btn = document.getElementById('createEventBtn');
  setButtonLoading(btn, true);
  try {
    const rawDate = document.getElementById('evDate').value;
    const ev = await api.createEvent({
      title: document.getElementById('evTitle').value,
      description: document.getElementById('evDescription').value,
      capacity: parseInt(document.getElementById('evCapacity').value, 10),
      event_date: new Date(rawDate).toISOString(),
    });
    alert(`Event "${ev.title}" created successfully!`);
    window.location.href = `./event.html?id=${ev.id}`;
  } catch (err) { toast(err.message, 'error'); }
  finally { setButtonLoading(btn, false); }
}
