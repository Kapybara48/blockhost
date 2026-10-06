/**
 * BlockHost - Minecraft Server Management Client
 * Autor: Daniel Krejčí
 * Webová aplikace (HTML, CSS, JS)
 */

// ==========================================================================
// 1. STATE & DEFAULT CONFIGURATION
// ==========================================================================
const STORAGE_KEY = 'blockhost_state_v1';

const DEFAULT_STATE = {
  serverStatus: 'online', // 'online' | 'offline' | 'restarting'
  uptimeSeconds: 14320,
  activeProfileId: 'standard',
  metrics: {
    cpu: 28,
    ramUsed: 3.2,
    tps: 20.0
  },
  activeEvent: null, // { name, profileId, endsAt, originalProfileId }
  profiles: [
    {
      id: 'eco',
      name: 'Eco Starter',
      ram: 2,
      vcpu: 1,
      price: 19,
      description: 'Úsporný režim pro noční provoz a minimální zátěž'
    },
    {
      id: 'standard',
      name: 'Standard Community',
      ram: 6,
      vcpu: 2,
      price: 49,
      description: 'Optimální profil pro běžné denní hraní a základní pluginy'
    },
    {
      id: 'pro',
      name: 'Pro Guild',
      ram: 10,
      vcpu: 4,
      price: 89,
      description: 'Vysoký výkon pro větší komunitu a náročnější modifikace'
    },
    {
      id: 'ultra',
      name: 'Ultra Event',
      ram: 16,
      vcpu: 6,
      price: 149,
      description: 'Maximální výkon pro velké herní eventy, minihry a turnaje'
    }
  ],
  schedules: [
    {
      id: 'sch-1',
      name: 'Noční útlum',
      days: ['Po', 'Út', 'St', 'Čt', 'Pá'],
      time: '01:00',
      targetProfileId: 'eco',
      enabled: true
    },
    {
      id: 'sch-2',
      name: 'Odpolední špička',
      days: ['Po', 'Út', 'St', 'Čt', 'Pá'],
      time: '14:30',
      targetProfileId: 'standard',
      enabled: true
    },
    {
      id: 'sch-3',
      name: 'Víkendový boost',
      days: ['So', 'Ne'],
      time: '10:00',
      targetProfileId: 'pro',
      enabled: true
    }
  ],
  players: [
    { name: 'DanKrejci', ip: '192.168.1.15', ping: 18, op: true, joined: '10:14', online: true },
    { name: 'MinerBob', ip: '178.40.12.91', ping: 32, op: false, joined: '11:05', online: true },
    { name: 'PixelQueen', ip: '89.102.4.21', ping: 24, op: false, joined: '11:22', online: true },
    { name: 'SteveMaster', ip: '37.188.19.4', ping: 45, op: false, joined: '11:45', online: true }
  ],
  whitelist: [
    { name: 'DanKrejci', addedDate: '2026-10-01', addedBy: 'CONSOLE' },
    { name: 'MinerBob', addedDate: '2026-10-01', addedBy: 'DanKrejci' },
    { name: 'PixelQueen', addedDate: '2026-10-02', addedBy: 'DanKrejci' },
    { name: 'SteveMaster', addedDate: '2026-10-02', addedBy: 'DanKrejci' },
    { name: 'AlexCraft', addedDate: '2026-10-03', addedBy: 'DanKrejci' },
    { name: 'RedstoneWiz', addedDate: '2026-10-04', addedBy: 'DanKrejci' }
  ],
  banlist: [
    { name: 'Griefer99', reason: 'Ničení cizích staveb (Griefing)', date: '2026-10-03', banner: 'DanKrejci' },
    { name: 'XrayUser', reason: 'Použití nepovolených módů (X-Ray)', date: '2026-10-04', banner: 'CONSOLE' }
  ],
  config: {
    motd: 'Vítejte na BlockHost Minecraft Serveru!',
    serverName: 'BlockHost SMP',
    core: 'Paper 1.20.4',
    gamemode: 'survival',
    difficulty: 'normal',
    maxPlayers: 40,
    port: 25565,
    pvp: true,
    whitelist: true
  },
  backups: [
    { id: 'bk-1', name: 'Auto-Zaloha-Svet-Daily', date: '2026-10-06 04:00', size: '428 MB', type: 'Automatická' },
    { id: 'bk-2', name: 'Pred-Aktualizaci-Paper', date: '2026-10-05 18:30', size: '415 MB', type: 'Manuální' },
    { id: 'bk-3', name: 'Tydenni-Archiv-Full', date: '2026-10-01 02:00', size: '610 MB', type: 'Automatická' }
  ],
  logs: [
    { time: '11:58:12', level: 'INFO', text: '[Server] Starting Minecraft server version 1.20.4' },
    { time: '11:58:15', level: 'INFO', text: '[Server] Loading properties and world chunks...' },
    { time: '11:58:20', level: 'INFO', text: '[BlockHost] Dynamic profile active: Standard (6 GB RAM, 2 vCPU)' },
    { time: '11:58:22', level: 'INFO', text: '[Server] Done (7.21s)! For help, type "help"' },
    { time: '12:01:05', level: 'CHAT', text: '<DanKrejci> Vítejte na serveru BlockHost!' },
    { time: '12:05:40', level: 'INFO', text: '[BlockHost-RCON] RCON daemon running on port 25575' }
  ]
};

// Application State container
let state = loadState();

function loadState() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      return { ...DEFAULT_STATE, ...JSON.parse(saved) };
    }
  } catch (e) {
    console.error('Chyba při načítání stavu z localStorage:', e);
  }
  return JSON.parse(JSON.stringify(DEFAULT_STATE));
}

function saveState() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } catch (e) {
    console.error('Chyba při ukládání stavu:', e);
  }
}

// ==========================================================================
// 2. INITIALIZATION & NAVIGATION
// ==========================================================================
document.addEventListener('DOMContentLoaded', () => {
  setupNavigation();
  setupEventListeners();
  renderAllViews();
  startSimulationTicker();
  updateRawConfigPreview();
});

function setupNavigation() {
  const navButtons = document.querySelectorAll('.nav-btn');
  navButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      const viewId = btn.getAttribute('data-view');
      switchView(viewId);
    });
  });

  const mobileToggle = document.getElementById('mobile-nav-toggle');
  const sidebar = document.getElementById('sidebar');
  if (mobileToggle) {
    mobileToggle.addEventListener('click', () => {
      sidebar.classList.toggle('open');
    });
  }
}

function switchView(viewId) {
  // Update nav buttons
  document.querySelectorAll('.nav-btn').forEach(btn => {
    btn.classList.toggle('active', btn.getAttribute('data-view') === viewId);
  });

  // Update view sections
  document.querySelectorAll('.view-section').forEach(sec => {
    sec.classList.toggle('active', sec.id === `view-${viewId}`);
  });

  // Close mobile sidebar if open
  const sidebar = document.getElementById('sidebar');
  if (sidebar) sidebar.classList.remove('open');

  // Specific triggers
  if (viewId === 'console') {
    scrollConsoleToBottom();
  }
}

// ==========================================================================
// 3. RENDERERS
// ==========================================================================
function renderAllViews() {
  renderHeaderAndMetrics();
  renderProfiles();
  renderScheduler();
  renderPlayers();
  renderConfigForm();
  renderBackups();
  renderConsoleLogs();
  checkActiveEvent();
}

// Render Header & Top Metrics
function renderHeaderAndMetrics() {
  const currentProfile = state.profiles.find(p => p.id === state.activeProfileId) || state.profiles[1];
  
  // Status pill
  const statusPill = document.getElementById('server-status-pill');
  const statusText = document.getElementById('server-status-text');
  statusPill.className = `status-indicator ${state.serverStatus}`;
  statusText.textContent = state.serverStatus.toUpperCase();

  // Control buttons state
  const btnStart = document.getElementById('btn-start');
  const btnStop = document.getElementById('btn-stop');
  const btnRestart = document.getElementById('btn-restart');
  
  if (state.serverStatus === 'online') {
    btnStart.disabled = true;
    btnStop.disabled = false;
    btnRestart.disabled = false;
  } else if (state.serverStatus === 'offline') {
    btnStart.disabled = false;
    btnStop.disabled = true;
    btnRestart.disabled = true;
  } else {
    // Restarting
    btnStart.disabled = true;
    btnStop.disabled = true;
    btnRestart.disabled = true;
  }

  // Header meta
  document.getElementById('header-version').textContent = state.config.core;
  document.getElementById('header-profile').textContent = `${currentProfile.name} (${currentProfile.ram} GB)`;
  document.getElementById('header-uptime').textContent = formatSeconds(state.uptimeSeconds);
  document.getElementById('sidebar-server-name').textContent = state.config.serverName;

  // Nav badges
  document.getElementById('nav-active-profile-badge').textContent = currentProfile.name.split(' ')[0];
  const onlinePlayersCount = state.serverStatus === 'online' ? state.players.filter(p => p.online).length : 0;
  document.getElementById('nav-players-count').textContent = onlinePlayersCount;

  // Dashboard Metrics
  const cpuVal = state.serverStatus === 'online' ? state.metrics.cpu : 0;
  const ramVal = state.serverStatus === 'online' ? state.metrics.ramUsed.toFixed(1) : '0.0';
  const tpsVal = state.serverStatus === 'online' ? state.metrics.tps.toFixed(1) : '0.0';
  const ramPercent = Math.min(100, Math.round((parseFloat(ramVal) / currentProfile.ram) * 100));

  document.getElementById('stat-cpu').textContent = `${cpuVal}%`;
  document.getElementById('bar-cpu').style.width = `${cpuVal}%`;
  document.getElementById('stat-vcpu-count').textContent = `${currentProfile.vcpu} vCPU dedikováno`;

  document.getElementById('stat-ram').textContent = `${ramVal} / ${currentProfile.ram} GB`;
  document.getElementById('bar-ram').style.width = `${ramPercent}%`;
  document.getElementById('stat-ram-percent').textContent = `${ramPercent}% využito`;
  document.getElementById('stat-ram-profile').textContent = `Profil: ${currentProfile.name}`;

  document.getElementById('stat-players').textContent = `${onlinePlayersCount} / ${state.config.maxPlayers}`;
  const playerPercent = Math.round((onlinePlayersCount / state.config.maxPlayers) * 100);
  document.getElementById('bar-players').style.width = `${playerPercent}%`;
  document.getElementById('stat-players-detail').textContent = `${onlinePlayersCount} online`;

  document.getElementById('stat-tps').textContent = tpsVal;
  const tpsPercent = Math.min(100, (parseFloat(tpsVal) / 20.0) * 100);
  document.getElementById('bar-tps').style.width = `${tpsPercent}%`;

  // Dashboard Active Profile Card
  document.getElementById('dash-profile-name').textContent = currentProfile.name;
  document.getElementById('dash-profile-desc').textContent = currentProfile.description;
  document.getElementById('dash-hourly-rate').textContent = `${currentProfile.price} Kč / hodina`;
  const estimatedMonth = Math.round(currentProfile.price * 24 * 30 * 0.7); // estimate with nighttime reduction
  document.getElementById('dash-monthly-cost').textContent = `~${estimatedMonth} Kč / měsíc (s plánovačem)`;
}

// Render Profiles View
function renderProfiles() {
  const container = document.getElementById('profiles-container');
  if (!container) return;

  const selectEventProfile = document.getElementById('event-select-profile');
  const selectScheduleProfile = document.getElementById('sch-rule-profile');

  container.innerHTML = '';
  if (selectEventProfile) selectEventProfile.innerHTML = '';
  if (selectScheduleProfile) selectScheduleProfile.innerHTML = '';

  state.profiles.forEach(p => {
    const isActive = p.id === state.activeProfileId;
    const card = document.createElement('div');
    card.className = `profile-card ${isActive ? 'active-profile' : ''}`;

    card.innerHTML = `
      ${isActive ? '<span class="active-profile-badge">Aktivní Profil</span>' : ''}
      <div class="profile-header">
        <h3 class="profile-name">${p.name}</h3>
        <p class="profile-desc">${p.description}</p>
      </div>
      <div class="profile-price">
        <span class="price-amount">${p.price} Kč</span>
        <span class="price-unit">/ hodina</span>
      </div>
      <ul class="profile-specs">
        <li>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 12h-4l-3 9L9 3l-3 9H2"></path></svg>
          <strong>${p.ram} GB</strong> RAM paměť
        </li>
        <li>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="4" width="16" height="16" rx="2"></rect><rect x="9" y="9" width="6" height="6"></rect></svg>
          <strong>${p.vcpu} vCPU</strong> dedikovaná jádra
        </li>
        <li>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
          Automatické NVMe zálohování
        </li>
        <li>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
          Dynamické přepínání bez výpadku
        </li>
      </ul>
      <div style="margin-top: auto;">
        ${isActive 
          ? `<button class="btn btn-secondary" style="width: 100%; justify-content: center;" disabled>Aktuálně aktivní</button>` 
          : `<button class="btn btn-primary" style="width: 100%; justify-content: center;" onclick="activateProfile('${p.id}')">Přepnout na tento profil</button>`
        }
      </div>
    `;

    container.appendChild(card);

    // Populate dropdowns
    if (selectEventProfile) {
      const opt = document.createElement('option');
      opt.value = p.id;
      opt.textContent = `${p.name} (${p.ram} GB RAM, ${p.price} Kč/h)`;
      if (p.id === 'ultra') opt.selected = true;
      selectEventProfile.appendChild(opt);
    }

    if (selectScheduleProfile) {
      const opt = document.createElement('option');
      opt.value = p.id;
      opt.textContent = `${p.name} (${p.ram} GB RAM)`;
      selectScheduleProfile.appendChild(opt);
    }
  });
}

// Activate Profile
function activateProfile(profileId, isAutomatic = false, eventTrigger = false) {
  const target = state.profiles.find(p => p.id === profileId);
  if (!target) return;

  state.activeProfileId = profileId;
  saveState();
  renderAllViews();

  addLog('INFO', `[BlockHost] Profil změněn na: ${target.name} (${target.ram} GB RAM, ${target.vcpu} vCPU)`);
  if (!isAutomatic) {
    showToast(`Profil byl úspěšně přepnut na: ${target.name}`);
  }
}

// Render Scheduler View
function renderScheduler() {
  const tbody = document.getElementById('schedule-table-body');
  if (!tbody) return;

  tbody.innerHTML = '';
  if (state.schedules.length === 0) {
    tbody.innerHTML = `<tr><td colspan="6" style="text-align: center; color: var(--text-muted); padding: 24px;">Žádná aktivní pravidla plánovače.</td></tr>`;
    return;
  }

  state.schedules.forEach(item => {
    const profile = state.profiles.find(p => p.id === item.targetProfileId) || { name: item.targetProfileId };
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><strong>${item.name}</strong></td>
      <td>
        ${item.days.map(d => `<span class="day-badge">${d}</span>`).join(' ')}
      </td>
      <td><strong>${item.time}</strong></td>
      <td><span style="color: var(--accent-green); font-weight: 600;">${profile.name}</span></td>
      <td>
        <label style="cursor: pointer;">
          <input type="checkbox" ${item.enabled ? 'checked' : ''} onchange="toggleSchedule('${item.id}', this.checked)">
          <span style="font-size: 0.8rem; margin-left: 6px; color: ${item.enabled ? 'var(--accent-green)' : 'var(--text-muted)'};">
            ${item.enabled ? 'Aktivní' : 'Vypnuto'}
          </span>
        </label>
      </td>
      <td style="text-align: right;">
        <button class="btn btn-secondary btn-sm" onclick="deleteSchedule('${item.id}')">Smazat</button>
      </td>
    `;
    tbody.appendChild(tr);
  });
}

function toggleSchedule(id, enabled) {
  const rule = state.schedules.find(s => s.id === id);
  if (rule) {
    rule.enabled = enabled;
    saveState();
    renderScheduler();
    showToast(`Pravidlo "${rule.name}" bylo ${enabled ? 'aktivováno' : 'deaktivováno'}.`);
  }
}

function deleteSchedule(id) {
  state.schedules = state.schedules.filter(s => s.id !== id);
  saveState();
  renderScheduler();
  showToast('Pravidlo plánovače bylo odstraněno.');
}

// Render Players View
function renderPlayers() {
  const onlineTbody = document.getElementById('online-players-tbody');
  const whitelistTbody = document.getElementById('whitelist-tbody');
  const banlistTbody = document.getElementById('banlist-tbody');

  // Counts
  const onlineCount = state.serverStatus === 'online' ? state.players.filter(p => p.online).length : 0;
  document.getElementById('count-online-tab').textContent = onlineCount;
  document.getElementById('count-whitelist-tab').textContent = state.whitelist.length;
  document.getElementById('count-banlist-tab').textContent = state.banlist.length;

  // Online Players
  if (onlineTbody) {
    onlineTbody.innerHTML = '';
    const activePlayers = state.serverStatus === 'online' ? state.players.filter(p => p.online) : [];
    if (activePlayers.length === 0) {
      onlineTbody.innerHTML = `<tr><td colspan="6" style="text-align: center; color: var(--text-muted); padding: 24px;">Žádní hráči nejsou online.</td></tr>`;
    } else {
      activePlayers.forEach(player => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td>
            <div style="display: flex; align-items: center; gap: 10px;">
              <div style="width: 28px; height: 28px; border-radius: 4px; background: #3b82f6; display: flex; align-items: center; justify-content: center; font-weight: bold; font-size: 0.75rem; color: white;">
                ${player.name.substring(0, 2).toUpperCase()}
              </div>
              <strong>${player.name}</strong>
            </div>
          </td>
          <td style="color: var(--text-muted);">${player.ip}</td>
          <td><span style="color: ${player.ping < 40 ? 'var(--accent-green)' : 'var(--accent-amber)'};">${player.ping} ms</span></td>
          <td>${player.op ? '<span style="color: #facc15; font-weight: 700;">OP Admin</span>' : '<span style="color: var(--text-secondary);">Hráč</span>'}</td>
          <td style="color: var(--text-muted);">${player.joined}</td>
          <td style="text-align: right;">
            <button class="btn btn-secondary btn-sm" onclick="togglePlayerOp('${player.name}')">${player.op ? 'Odebrat OP' : 'Dát OP'}</button>
            <button class="btn btn-secondary btn-sm" onclick="kickPlayer('${player.name}')">Kick</button>
            <button class="btn btn-danger btn-sm" onclick="banPlayerDirect('${player.name}')">Ban</button>
          </td>
        `;
        onlineTbody.appendChild(tr);
      });
    }
  }

  // Whitelist
  if (whitelistTbody) {
    whitelistTbody.innerHTML = '';
    state.whitelist.forEach(item => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td><strong>${item.name}</strong></td>
        <td style="color: var(--text-muted);">${item.addedDate}</td>
        <td><span style="color: var(--accent-green);">${item.addedBy}</span></td>
        <td style="text-align: right;">
          <button class="btn btn-secondary btn-sm" onclick="removeWhitelist('${item.name}')">Odebrat</button>
        </td>
      `;
      whitelistTbody.appendChild(tr);
    });
  }

  // Banlist
  if (banlistTbody) {
    banlistTbody.innerHTML = '';
    if (state.banlist.length === 0) {
      banlistTbody.innerHTML = `<tr><td colspan="5" style="text-align: center; color: var(--text-muted); padding: 24px;">Banlist je prázdný.</td></tr>`;
    } else {
      state.banlist.forEach(item => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td><strong style="color: var(--accent-red);">${item.name}</strong></td>
          <td>${item.reason}</td>
          <td style="color: var(--text-muted);">${item.date}</td>
          <td>${item.banner}</td>
          <td style="text-align: right;">
            <button class="btn btn-secondary btn-sm" onclick="unbanPlayer('${item.name}')">Zrušit BAN (Unban)</button>
          </td>
        `;
        banlistTbody.appendChild(tr);
      });
    }
  }
}

function switchPlayerTab(tab) {
  document.getElementById('tab-btn-online').classList.toggle('active-tab', tab === 'online');
  document.getElementById('tab-btn-whitelist').classList.toggle('active-tab', tab === 'whitelist');
  document.getElementById('tab-btn-banlist').classList.toggle('active-tab', tab === 'banlist');

  document.getElementById('players-table-container').style.display = tab === 'online' ? 'block' : 'none';
  document.getElementById('whitelist-table-container').style.display = tab === 'whitelist' ? 'block' : 'none';
  document.getElementById('banlist-table-container').style.display = tab === 'banlist' ? 'block' : 'none';
}

function kickPlayer(name) {
  const p = state.players.find(x => x.name === name);
  if (p) {
    p.online = false;
    addLog('INFO', `[Server] Kicked player ${name} from the server (by admin)`);
    saveState();
    renderAllViews();
    showToast(`Hráč ${name} byl vyhozen ze serveru.`);
  }
}

function banPlayerDirect(name) {
  kickPlayer(name);
  if (!state.banlist.find(b => b.name === name)) {
    state.banlist.push({
      name,
      reason: 'Zabanováno administrátorem',
      date: new Date().toISOString().split('T')[0],
      banner: 'Admin'
    });
    addLog('WARN', `[Server] Banned player ${name}: Zabanováno administrátorem`);
    saveState();
    renderAllViews();
    showToast(`Hráč ${name} byl zabanován.`, 'error');
  }
}

function unbanPlayer(name) {
  state.banlist = state.banlist.filter(b => b.name !== name);
  addLog('INFO', `[Server] Unbanned player ${name}`);
  saveState();
  renderAllViews();
  showToast(`Hráč ${name} byl odbanován.`);
}

function togglePlayerOp(name) {
  const p = state.players.find(x => x.name === name);
  if (p) {
    p.op = !p.op;
    addLog('INFO', `[Server] ${p.op ? 'Opped' : 'De-opped'} player ${name}`);
    saveState();
    renderPlayers();
    showToast(`Oprávnění OP pro hráče ${name} bylo ${p.op ? 'uděleno' : 'odebráno'}.`);
  }
}

function removeWhitelist(name) {
  state.whitelist = state.whitelist.filter(w => w.name !== name);
  addLog('INFO', `[Server] Removed ${name} from whitelist`);
  saveState();
  renderPlayers();
  showToast(`Hráč ${name} byl odebrán z whitelistu.`);
}

// Render Backups View
function renderBackups() {
  const tbody = document.getElementById('backups-tbody');
  if (!tbody) return;

  tbody.innerHTML = '';
  if (state.backups.length === 0) {
    tbody.innerHTML = `<tr><td colspan="5" style="text-align: center; color: var(--text-muted); padding: 24px;">Zatím nebyly vytvořeny žádné zálohy.</td></tr>`;
    return;
  }

  state.backups.forEach(bk => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><strong>${bk.name}</strong></td>
      <td style="color: var(--text-muted);">${bk.date}</td>
      <td><strong>${bk.size}</strong></td>
      <td><span class="day-badge" style="color: var(--accent-green);">${bk.type}</span></td>
      <td style="text-align: right;">
        <button class="btn btn-warning btn-sm" onclick="promptRestoreBackup('${bk.id}')">Obnovit</button>
        <button class="btn btn-secondary btn-sm" onclick="downloadBackup('${bk.id}')">Stáhnout</button>
        <button class="btn btn-secondary btn-sm" onclick="deleteBackup('${bk.id}')">Smazat</button>
      </td>
    `;
    tbody.appendChild(tr);
  });
}

// Render Configuration Form
function renderConfigForm() {
  const cfg = state.config;
  document.getElementById('cfg-motd').value = cfg.motd;
  document.getElementById('cfg-name').value = cfg.serverName;
  document.getElementById('cfg-core').value = cfg.core;
  document.getElementById('cfg-gamemode').value = cfg.gamemode;
  document.getElementById('cfg-difficulty').value = cfg.difficulty;
  document.getElementById('cfg-max-players').value = cfg.maxPlayers;
  document.getElementById('cfg-port').value = cfg.port;
  document.getElementById('cfg-pvp').checked = cfg.pvp;
  document.getElementById('cfg-whitelist').checked = cfg.whitelist;
}

function updateRawConfigPreview() {
  const cfg = state.config;
  const rawText = `# Minecraft server properties - BlockHost Generated
# ${new Date().toISOString()}
motd=${cfg.motd}
server-name=${cfg.serverName}
server-port=${cfg.port}
gamemode=${cfg.gamemode}
difficulty=${cfg.difficulty}
max-players=${cfg.maxPlayers}
pvp=${cfg.pvp}
white-list=${cfg.whitelist}
view-distance=10
simulation-distance=8
enable-rcon=true
rcon.port=25575
rcon.password=********
online-mode=true
allow-nether=true
enable-command-block=true
`;
  const preview = document.getElementById('raw-config-preview');
  if (preview) preview.textContent = rawText;
}

// Render Console Logs
function renderConsoleLogs() {
  const container = document.getElementById('terminal-logs-container');
  const miniContainer = document.getElementById('dashboard-mini-logs');
  if (!container) return;

  const html = state.logs.map(log => {
    let cls = 'log-info';
    if (log.level === 'WARN') cls = 'log-warn';
    if (log.level === 'ERROR') cls = 'log-error';
    if (log.level === 'CHAT') cls = 'log-chat';
    if (log.level === 'SYSTEM') cls = 'log-system';
    return `<div class="${cls}">[${log.time}] [${log.level}]: ${escapeHtml(log.text)}</div>`;
  }).join('');

  container.innerHTML = html;
  if (miniContainer) {
    miniContainer.innerHTML = html;
    miniContainer.scrollTop = miniContainer.scrollHeight;
  }
}

function scrollConsoleToBottom() {
  const container = document.getElementById('terminal-logs-container');
  if (container) container.scrollTop = container.scrollHeight;
}

function addLog(level, text) {
  const now = new Date();
  const time = now.toTimeString().split(' ')[0];
  state.logs.push({ time, level, text });
  if (state.logs.length > 200) {
    state.logs.shift();
  }
  renderConsoleLogs();
  scrollConsoleToBottom();
}

function clearConsole() {
  state.logs = [];
  renderConsoleLogs();
  showToast('Konzole byla vymazána.');
}

// ==========================================================================
// 4. EVENT BOOSTER / SCHEDULER LOGIC
// ==========================================================================
function checkActiveEvent() {
  const banner = document.getElementById('event-banner');
  const eventDetails = document.getElementById('event-active-details');
  const eventBadge = document.getElementById('event-badge-status');

  if (!state.activeEvent) {
    if (banner) banner.style.display = 'none';
    if (eventDetails) eventDetails.style.display = 'none';
    if (eventBadge) {
      eventBadge.className = 'status-indicator online';
      eventBadge.textContent = 'Žádný aktivní event';
    }
    return;
  }

  const now = Date.now();
  const diffMs = state.activeEvent.endsAt - now;

  if (diffMs <= 0) {
    // Event finished!
    const finishedEvent = state.activeEvent;
    state.activeEvent = null;
    activateProfile(finishedEvent.originalProfileId, true);
    addLog('SYSTEM', `[BlockHost Event] Event "${finishedEvent.name}" skončil. Profil automaticky vrácen na výchozí.`);
    showToast(`Event "${finishedEvent.name}" skončil. Server se vrátil na původní profil.`, 'warning');
    saveState();
    renderAllViews();
    return;
  }

  // Active event display
  const minsLeft = Math.floor(diffMs / 60000);
  const secsLeft = Math.floor((diffMs % 60000) / 1000);
  const timeStr = `${String(minsLeft).padStart(2, '0')}:${String(secsLeft).padStart(2, '0')}`;

  const prof = state.profiles.find(p => p.id === state.activeEvent.profileId) || { name: 'Ultra', ram: 16 };

  if (banner) {
    banner.style.display = 'flex';
    document.getElementById('event-banner-text').innerHTML = `<strong>${state.activeEvent.name}:</strong> Aktivní profil <strong>${prof.name} (${prof.ram} GB RAM)</strong>.`;
    document.getElementById('event-timer').textContent = timeStr;
  }

  if (eventDetails) {
    eventDetails.style.display = 'block';
    document.getElementById('event-active-name').textContent = state.activeEvent.name;
    document.getElementById('event-active-profile').textContent = `Profil: ${prof.name} (${prof.ram} GB RAM, ${prof.price} Kč/h)`;
    document.getElementById('event-active-time-left').textContent = timeStr;
  }

  if (eventBadge) {
    eventBadge.className = 'status-indicator online';
    eventBadge.textContent = 'PROBÍHÁ EVENT';
  }
}

// ==========================================================================
// 5. SERVER CONTROL ACTIONS
// ==========================================================================
function startServer() {
  if (state.serverStatus === 'online') return;
  state.serverStatus = 'restarting';
  renderHeaderAndMetrics();
  addLog('INFO', '[BlockHost] Starting server process...');
  showToast('Spouštím Minecraft server...');

  setTimeout(() => {
    state.serverStatus = 'online';
    state.uptimeSeconds = 0;
    addLog('INFO', '[Server] Minecraft server loaded chunks and opened network socket (25565)');
    addLog('INFO', `[Server] BlockHost profile: ${state.activeProfileId.toUpperCase()}`);
    saveState();
    renderAllViews();
    showToast('Server úspěšně běží online!', 'success');
  }, 2200);
}

function stopServer() {
  if (state.serverStatus === 'offline') return;
  state.serverStatus = 'offline';
  state.metrics.cpu = 0;
  state.metrics.ramUsed = 0;
  state.metrics.tps = 0;
  addLog('WARN', '[Server] Stopping server: saving players, saving worlds');
  addLog('INFO', '[Server] Thread shutdown complete.');
  saveState();
  renderAllViews();
  showToast('Server byl bezpečně zastaven.', 'warning');
}

function restartServer() {
  if (state.serverStatus === 'restarting') return;
  state.serverStatus = 'restarting';
  renderHeaderAndMetrics();
  addLog('WARN', '[Server] Restarting Minecraft server process...');
  showToast('Restartuji Minecraft server...');

  setTimeout(() => {
    state.serverStatus = 'online';
    state.uptimeSeconds = 0;
    addLog('INFO', '[Server] Minecraft server restarted and ready.');
    saveState();
    renderAllViews();
    showToast('Server byl úspěšně restartován!', 'success');
  }, 2800);
}

// ==========================================================================
// 6. BACKUP ENGINE
// ==========================================================================
function createBackup() {
  const progressCard = document.getElementById('backup-progress-card');
  const progressBar = document.getElementById('backup-progress-bar');
  const progressText = document.getElementById('backup-progress-text');

  progressCard.style.display = 'block';
  let percent = 0;
  progressBar.style.width = '0%';
  progressText.textContent = '0%';
  addLog('INFO', '[BlockHost Backup] Zahajuji zálohování světa a konfigurace...');

  const interval = setInterval(() => {
    percent += 15;
    if (percent > 100) percent = 100;
    progressBar.style.width = `${percent}%`;
    progressText.textContent = `${percent}%`;

    if (percent === 100) {
      clearInterval(interval);
      setTimeout(() => {
        progressCard.style.display = 'none';
        const now = new Date();
        const dateStr = now.toISOString().replace('T', ' ').substring(0, 16);
        const newBackup = {
          id: `bk-${Date.now()}`,
          name: `Manual-Zaloha-${now.toLocaleDateString('cs-CZ').replace(/\./g, '-')}`,
          date: dateStr,
          size: `${Math.floor(400 + Math.random() * 80)} MB`,
          type: 'Manuální'
        };
        state.backups.unshift(newBackup);
        addLog('INFO', `[BlockHost Backup] Záloha "${newBackup.name}" úspěšně vytvořena.`);
        saveState();
        renderBackups();
        showToast(`Záloha "${newBackup.name}" byla úspěšně dokončena!`);
      }, 500);
    }
  }, 200);
}

let pendingRestoreId = null;

function promptRestoreBackup(id) {
  const bk = state.backups.find(b => b.id === id);
  if (!bk) return;
  pendingRestoreId = id;
  document.getElementById('restore-backup-desc').textContent = 
    `Opravdu si přejete obnovit server ze zálohy "${bk.name}" (${bk.date}, ${bk.size})? Aktuální data budou nahrazena.`;
  openModal('modal-restore-backup');
}

function executeRestoreBackup() {
  closeModal('modal-restore-backup');
  if (!pendingRestoreId) return;
  const bk = state.backups.find(b => b.id === pendingRestoreId);
  pendingRestoreId = null;

  if (bk) {
    showToast(`Obnovuji data ze zálohy ${bk.name}...`);
    addLog('WARN', `[BlockHost Backup] Obnova dat ze zálohy: ${bk.name}`);
    restartServer();
  }
}

function downloadBackup(id) {
  const bk = state.backups.find(b => b.id === id);
  if (!bk) return;

  const dummyContent = `BlockHost Minecraft Backup Archive\nName: ${bk.name}\nDate: ${bk.date}\nSize: ${bk.size}\nStatus: Verified\n`;
  const blob = new Blob([dummyContent], { type: 'text/plain' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${bk.name}.tar.gz`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
  showToast(`Stahování archivu ${bk.name}.tar.gz zahájeno.`);
}

function deleteBackup(id) {
  state.backups = state.backups.filter(b => b.id !== id);
  saveState();
  renderBackups();
  showToast('Záloha byla odstraněna.');
}

// ==========================================================================
// 7. CONSOLE COMMAND PARSER
// ==========================================================================
function sendConsoleCommand(cmdString) {
  if (!cmdString || !cmdString.trim()) return;
  cmdString = cmdString.trim();

  addLog('INFO', `> ${cmdString}`);

  const parts = cmdString.split(' ');
  const cmd = parts[0].toLowerCase().replace(/^\//, ''); // allow both /help and help
  const args = parts.slice(1);

  if (state.serverStatus !== 'online' && !['start', 'help'].includes(cmd)) {
    addLog('ERROR', '[RCON] Příkaz nelze vykonat: Server je offline.');
    return;
  }

  switch (cmd) {
    case 'help':
      addLog('SYSTEM', 'Dostupné příkazy:');
      addLog('SYSTEM', '  help                 - Zobrazí nápovědu');
      addLog('SYSTEM', '  list                 - Seznam připojených hráčů');
      addLog('SYSTEM', '  tps                  - Zobrazí aktuální TPS');
      addLog('SYSTEM', '  say <zpráva>         - Odešle hlášení do chatu všem hráčům');
      addLog('SYSTEM', '  op <hráč>            - Přidělí hráči administrátorská práva');
      addLog('SYSTEM', '  deop <hráč>          - Odebere hráči OP práva');
      addLog('SYSTEM', '  kick <hráč>          - Vyhodí hráče ze serveru');
      addLog('SYSTEM', '  ban <hráč>           - Zabanuje hráče');
      addLog('SYSTEM', '  unban <hráč>         - Odbanuje hráče');
      addLog('SYSTEM', '  profile <eco|standard|pro|ultra> - Okamžitě přepne výkonový profil');
      addLog('SYSTEM', '  backup               - Spustí okamžité zálohování');
      addLog('SYSTEM', '  start / stop / restart - Řízení stavu serveru');
      addLog('SYSTEM', '  clear                - Vymaže okno konzole');
      break;

    case 'list': {
      const active = state.players.filter(p => p.online);
      addLog('INFO', `Online hráči (${active.length}/${state.config.maxPlayers}): ${active.map(p => p.name).join(', ') || 'Žádní'}`);
      break;
    }

    case 'tps':
      addLog('INFO', `TPS z posledních 1m, 5m, 15m: ${state.metrics.tps.toFixed(2)}, 20.00, 20.00`);
      break;

    case 'say':
      if (args.length === 0) {
        addLog('ERROR', 'Použití: say <zpráva>');
      } else {
        const msg = args.join(' ');
        addLog('CHAT', `[Server] ${msg}`);
      }
      break;

    case 'op': {
      const nick = args[0];
      if (!nick) {
        addLog('ERROR', 'Použití: op <hráč>');
      } else {
        togglePlayerOp(nick);
      }
      break;
    }

    case 'deop': {
      const nick = args[0];
      if (!nick) {
        addLog('ERROR', 'Použití: deop <hráč>');
      } else {
        const p = state.players.find(x => x.name.toLowerCase() === nick.toLowerCase());
        if (p) {
          p.op = false;
          saveState();
          renderPlayers();
          addLog('INFO', `[Server] De-opped player ${p.name}`);
        } else {
          addLog('WARN', `Hráč ${nick} nebyl nalezen.`);
        }
      }
      break;
    }

    case 'kick': {
      const nick = args[0];
      if (!nick) {
        addLog('ERROR', 'Použití: kick <hráč>');
      } else {
        kickPlayer(nick);
      }
      break;
    }

    case 'ban': {
      const nick = args[0];
      if (!nick) {
        addLog('ERROR', 'Použití: ban <hráč>');
      } else {
        banPlayerDirect(nick);
      }
      break;
    }

    case 'unban': {
      const nick = args[0];
      if (!nick) {
        addLog('ERROR', 'Použití: unban <hráč>');
      } else {
        unbanPlayer(nick);
      }
      break;
    }

    case 'profile': {
      const pId = args[0]?.toLowerCase();
      if (!pId) {
        addLog('ERROR', 'Použití: profile <eco|standard|pro|ultra>');
      } else {
        activateProfile(pId);
      }
      break;
    }

    case 'backup':
      createBackup();
      break;

    case 'clear':
      clearConsole();
      break;

    case 'start':
      startServer();
      break;

    case 'stop':
      stopServer();
      break;

    case 'restart':
      restartServer();
      break;

    default:
      addLog('WARN', `Neznámý příkaz "${cmd}". Pro seznam příkazů napište "help".`);
  }
}

// ==========================================================================
// 8. SIMULATION TICKER (CPU/RAM fluctuation, Event timer, Scheduler check)
// ==========================================================================
function startSimulationTicker() {
  setInterval(() => {
    // 1. Event Timer Check
    if (state.activeEvent) {
      checkActiveEvent();
    }

    // 2. Metrics simulation if server is online
    if (state.serverStatus === 'online') {
      state.uptimeSeconds += 1;
      
      const currentProfile = state.profiles.find(p => p.id === state.activeProfileId) || state.profiles[1];
      const onlineCount = state.players.filter(p => p.online).length;

      // Realistic CPU calculation
      const baseCpu = 15 + (onlineCount * 3.5);
      const jitter = (Math.random() * 8) - 4;
      state.metrics.cpu = Math.max(5, Math.min(99, Math.round(baseCpu + jitter)));

      // Realistic RAM calculation
      const baseRam = 1.8 + (onlineCount * 0.35);
      const ramJitter = (Math.random() * 0.2) - 0.1;
      state.metrics.ramUsed = Math.min(currentProfile.ram * 0.95, Math.max(1.0, baseRam + ramJitter));

      // TPS calculation
      if (state.metrics.cpu > 90) {
        state.metrics.tps = Math.max(16.5, 20.0 - (Math.random() * 2));
      } else {
        state.metrics.tps = Math.min(20.0, 19.85 + (Math.random() * 0.15));
      }

      // Update UI metrics
      renderHeaderAndMetrics();
    }
  }, 1000);
}

// ==========================================================================
// 9. EVENT LISTENERS & MODALS
// ==========================================================================
function setupEventListeners() {
  // Top control buttons
  document.getElementById('btn-start').addEventListener('click', startServer);
  document.getElementById('btn-stop').addEventListener('click', stopServer);
  document.getElementById('btn-restart').addEventListener('click', restartServer);
  document.getElementById('btn-quick-refresh').addEventListener('click', () => {
    renderAllViews();
    showToast('Data byla aktualizována.');
  });

  // End active event button
  document.getElementById('btn-end-event').addEventListener('click', () => {
    if (state.activeEvent) {
      const orig = state.activeEvent.originalProfileId;
      state.activeEvent = null;
      activateProfile(orig, false);
      showToast('Herní event byl předčasně ukončen.', 'warning');
    }
  });

  // Save Config Button
  document.getElementById('btn-save-config').addEventListener('click', () => {
    state.config.motd = document.getElementById('cfg-motd').value;
    state.config.serverName = document.getElementById('cfg-name').value;
    state.config.core = document.getElementById('cfg-core').value;
    state.config.gamemode = document.getElementById('cfg-gamemode').value;
    state.config.difficulty = document.getElementById('cfg-difficulty').value;
    state.config.maxPlayers = parseInt(document.getElementById('cfg-max-players').value, 10) || 40;
    state.config.port = parseInt(document.getElementById('cfg-port').value, 10) || 25565;
    state.config.pvp = document.getElementById('cfg-pvp').checked;
    state.config.whitelist = document.getElementById('cfg-whitelist').checked;

    saveState();
    updateRawConfigPreview();
    renderHeaderAndMetrics();
    addLog('INFO', `[Server] Configuration saved: MOTD="${state.config.motd}", Core=${state.config.core}`);
    showToast('Konfigurace server.properties byla uložena!');
  });

  // Dynamic preview update while typing in config
  ['cfg-motd', 'cfg-name', 'cfg-core', 'cfg-gamemode', 'cfg-difficulty', 'cfg-max-players', 'cfg-port', 'cfg-pvp', 'cfg-whitelist'].forEach(id => {
    const el = document.getElementById(id);
    if (el) {
      el.addEventListener('input', () => {
        state.config[id.replace('cfg-', '')] = el.type === 'checkbox' ? el.checked : el.value;
        updateRawConfigPreview();
      });
    }
  });

  // Backup create button
  document.getElementById('btn-create-backup').addEventListener('click', createBackup);
  document.getElementById('btn-confirm-restore').addEventListener('click', executeRestoreBackup);

  // Terminal Console Form
  const consoleForm = document.getElementById('console-form');
  const consoleInput = document.getElementById('console-input');
  if (consoleForm && consoleInput) {
    consoleForm.addEventListener('submit', (e) => {
      e.preventDefault();
      const val = consoleInput.value;
      consoleInput.value = '';
      sendConsoleCommand(val);
    });
  }

  // Create Custom Profile Modal submit
  document.getElementById('btn-save-new-profile').addEventListener('click', () => {
    const name = document.getElementById('new-profile-name').value.trim();
    const desc = document.getElementById('new-profile-desc').value.trim();
    const ram = parseInt(document.getElementById('new-profile-ram').value, 10) || 4;
    const vcpu = parseInt(document.getElementById('new-profile-vcpu').value, 10) || 2;
    const price = parseInt(document.getElementById('new-profile-price').value, 10) || 35;

    if (!name) {
      alert('Zadejte prosím název profilu.');
      return;
    }

    const id = `custom-${Date.now()}`;
    state.profiles.push({ id, name, description: desc || 'Uživatelský profil', ram, vcpu, price });
    saveState();
    closeModal('modal-create-profile');
    renderProfiles();
    showToast(`Vlastní profil "${name}" byl vytvořen!`);
  });

  // Launch Event Modal submit
  document.getElementById('btn-confirm-launch-event').addEventListener('click', () => {
    const name = document.getElementById('event-input-name').value.trim() || 'Komunitní Turnaj';
    const profileId = document.getElementById('event-select-profile').value;
    const durationMins = parseInt(document.getElementById('event-duration').value, 10) || 60;

    const endsAt = Date.now() + (durationMins * 60 * 1000);
    const originalProfileId = state.activeProfileId;

    state.activeEvent = {
      name,
      profileId,
      endsAt,
      durationMins,
      originalProfileId
    };

    activateProfile(profileId, true);
    saveState();
    closeModal('modal-event');
    renderAllViews();
    addLog('SYSTEM', `[BlockHost Event] Spuštěn event "${name}" na dobu ${durationMins} minut. Dočasný profil: ${profileId}`);
    showToast(`Herní event "${name}" byl úspěšně spuštěn!`);
  });

  // Add Schedule Rule Modal submit
  document.getElementById('btn-save-schedule').addEventListener('click', () => {
    const name = document.getElementById('sch-rule-name').value.trim();
    const targetProfileId = document.getElementById('sch-rule-profile').value;
    const time = document.getElementById('sch-rule-time').value;
    const checkedDays = Array.from(document.querySelectorAll('input[name="sch-day"]:checked')).map(cb => cb.value);

    if (!name || checkedDays.length === 0) {
      alert('Vyplňte prosím název pravidla a vyberte alespoň jeden den.');
      return;
    }

    state.schedules.push({
      id: `sch-${Date.now()}`,
      name,
      days: checkedDays,
      time,
      targetProfileId,
      enabled: true
    });

    saveState();
    closeModal('modal-add-schedule');
    renderScheduler();
    showToast(`Pravidlo plánovače "${name}" bylo přidáno!`);
  });

  // Add Player Modal submit
  const addActionSelect = document.getElementById('player-add-action');
  const banReasonGroup = document.getElementById('ban-reason-group');
  if (addActionSelect) {
    addActionSelect.addEventListener('change', () => {
      banReasonGroup.style.display = addActionSelect.value === 'ban' ? 'block' : 'none';
    });
  }

  document.getElementById('btn-confirm-add-player').addEventListener('click', () => {
    const nick = document.getElementById('player-add-name').value.trim();
    const action = document.getElementById('player-add-action').value;
    const reason = document.getElementById('player-add-reason').value.trim();

    if (!nick) {
      alert('Zadejte herní přezdívku hráče.');
      return;
    }

    if (action === 'whitelist') {
      if (!state.whitelist.find(w => w.name.toLowerCase() === nick.toLowerCase())) {
        state.whitelist.push({
          name: nick,
          addedDate: new Date().toISOString().split('T')[0],
          addedBy: 'Admin'
        });
        addLog('INFO', `[Server] Added ${nick} to whitelist`);
        showToast(`Hráč ${nick} byl přidán na whitelist.`);
      }
    } else if (action === 'ban') {
      banPlayerDirect(nick);
    } else if (action === 'simulate_join') {
      if (state.serverStatus !== 'online') {
        alert('Server musí být online pro simulaci připojení hráče.');
        return;
      }
      let existing = state.players.find(p => p.name.toLowerCase() === nick.toLowerCase());
      if (existing) {
        existing.online = true;
      } else {
        state.players.push({
          name: nick,
          ip: `192.168.1.${Math.floor(20 + Math.random() * 80)}`,
          ping: Math.floor(15 + Math.random() * 35),
          op: false,
          joined: new Date().toTimeString().substring(0, 5),
          online: true
        });
      }
      addLog('INFO', `[Server] ${nick}[/192.168.1.55:54321] logged in with entity id ${Math.floor(100 + Math.random() * 900)}`);
      showToast(`Hráč ${nick} se připojil na server.`);
    }

    saveState();
    closeModal('modal-add-player');
    renderPlayers();
  });
}

// ==========================================================================
// 10. HELPER FUNCTIONS
// ==========================================================================
function openModal(modalId) {
  const modal = document.getElementById(modalId);
  if (modal) modal.classList.add('active');
}

function closeModal(modalId) {
  const modal = document.getElementById(modalId);
  if (modal) modal.classList.remove('active');
}

function showToast(message, type = 'success') {
  const container = document.getElementById('toast-container');
  if (!container) return;

  const toast = document.createElement('div');
  toast.className = `toast ${type === 'error' ? 'toast-error' : type === 'warning' ? 'toast-warning' : ''}`;
  toast.innerHTML = `
    <span>${type === 'error' ? '❌' : type === 'warning' ? '⚠️' : '✅'}</span>
    <span>${escapeHtml(message)}</span>
  `;

  container.appendChild(toast);
  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';
    toast.style.transition = 'all 0.3s ease';
    setTimeout(() => toast.remove(), 300);
  }, 3500);
}

function formatSeconds(seconds) {
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  return `${String(hrs).padStart(2, '0')}:${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

function escapeHtml(str) {
  if (!str) return '';
  return str.replace(/[&<>'"]/g, 
    tag => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[tag] || tag)
  );
}
