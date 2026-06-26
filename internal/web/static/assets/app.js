(() => {
  "use strict";

  const $ = (sel) => document.querySelector(sel);
  const $$ = (sel) => [...document.querySelectorAll(sel)];

  let allCases = [];
  let allTemplates = [];
  let activeTemplate = null;

  const ACCOUNT_KEY = "asylum_account_name";
  const ACCOUNT_KEY_LEGACY = "asylum_artist_name";
  const BANCAMP_KEY = "asylum_bancamp_profile";
  const EDIT_TOKEN_KEY = "asylum_edit_token";

  let editEnabled = false;
  let activeCase = null;

  const PLATFORM_URLS = {
    "Bandcamp": (n) => {
      const h = normalizeHandle(n);
      return h ? `https://${encodeURIComponent(h.toLowerCase())}.bandcamp.com` : `https://bandcamp.com/search?q=${encodeURIComponent(n)}`;
    },
    "Spotify": (n) => `https://open.spotify.com/search/${encodeURIComponent(n)}`,
    "YouTube": (n) => {
      const h = normalizeHandle(n);
      return h ? `https://www.youtube.com/@${encodeURIComponent(h)}` : `https://www.youtube.com/results?search_query=${encodeURIComponent(n)}`;
    },
    "TikTok": (n) => {
      const h = normalizeHandle(n);
      return h ? `https://www.tiktok.com/@${encodeURIComponent(h)}` : `https://www.tiktok.com/search?q=${encodeURIComponent(n)}`;
    },
    "Instagram": (n) => {
      const h = normalizeHandle(n);
      return h ? `https://www.instagram.com/${encodeURIComponent(h)}/` : `https://www.instagram.com/explore/search/keyword/?q=${encodeURIComponent(n)}`;
    },
    "Apple Music": (n) => `https://music.apple.com/search?term=${encodeURIComponent(n)}`,
    "SoundCloud": (n) => {
      const h = normalizeHandle(n);
      return h ? `https://soundcloud.com/${encodeURIComponent(h.toLowerCase())}` : `https://soundcloud.com/search?q=${encodeURIComponent(n)}`;
    },
    "DistroKid": (n) => `https://www.google.com/search?q=${encodeURIComponent("site:distrokid.com " + n)}`,
    "Facebook": (n) => `https://www.facebook.com/search/top?q=${encodeURIComponent(n)}`,
    "X / Twitter": (n) => {
      const h = normalizeHandle(n);
      return h ? `https://x.com/${encodeURIComponent(h)}` : `https://x.com/search?q=${encodeURIComponent(n)}&src=typed_query`;
    },
  };

  function normalizeHandle(name) {
    const trimmed = (name || "").trim().replace(/^@+/, "");
    return trimmed.includes(" ") ? "" : trimmed;
  }

  function buildProfileURL(platform, accountName) {
    if (!platform || !accountName) return "";
    const fn = PLATFORM_URLS[platform];
    return fn ? fn(accountName.trim()) : "";
  }

  function accountNameOf(c) {
    return c.account_name || c.artist_name || "Unknown account";
  }

  const els = {
    grid: $("#cases-grid"),
    templatesGrid: $("#templates-grid"),
    loading: $("#cases-loading"),
    empty: $("#cases-empty"),
    statTotal: $("#stat-total"),
    statPlatforms: $("#stat-platforms"),
    search: $("#search"),
    filterPlatform: $("#filter-platform"),
    filterIncident: $("#filter-incident"),
    sortBy: $("#sort-by"),
    caseModal: $("#case-modal"),
    signalModal: $("#signal-modal"),
    submitModal: $("#submit-modal"),
    signalForm: $("#signal-form"),
    submitForm: $("#submit-form"),
    formError: $("#form-error"),
    signalError: $("#signal-error"),
    toast: $("#toast"),
    fileDrop: $("#file-drop"),
    fileDropText: $("#file-drop-text"),
    proofInput: $("#proof_file"),
    submitBtn: $("#submit-btn"),
    signalBtn: $("#signal-btn"),
    signalFileDrop: $("#signal-file-drop"),
    signalFileDropText: $("#signal-file-drop-text"),
    signalProofInput: $("#signal_proof_file"),
    editModal: $("#edit-modal"),
    editForm: $("#edit-form"),
    editError: $("#edit-error"),
    editBtn: $("#edit-btn"),
    editFileDrop: $("#edit-file-drop"),
    editFileDropText: $("#edit-file-drop-text"),
    editProofInput: $("#edit_proof_file"),
    modalFooter: $("#modal-footer"),
  };

  function escapeHtml(str) {
    const d = document.createElement("div");
    d.textContent = str ?? "";
    return d.innerHTML;
  }

  function formatDate(iso) {
    try {
      return new Intl.DateTimeFormat("en-GB", {
        day: "numeric",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      }).format(new Date(iso));
    } catch {
      return iso;
    }
  }

  function truncate(str, len) {
    if (!str || str.length <= len) return str || "";
    return str.slice(0, len).trim() + "…";
  }

  function proofUrl(filename) {
    return `/api/proof/${encodeURIComponent(filename)}`;
  }

  function isImage(ext) {
    return [".png", ".jpg", ".jpeg", ".gif", ".webp"].includes(ext);
  }

  function getExt(filename) {
    const i = filename.lastIndexOf(".");
    return i >= 0 ? filename.slice(i).toLowerCase() : "";
  }

  function submissionLabel(c) {
    if (c.submission_type === "signal") return "Signal";
    return "Report";
  }

  function showToast(msg, type = "success") {
    els.toast.textContent = msg;
    els.toast.className = `toast ${type}`;
    els.toast.hidden = false;
    clearTimeout(showToast._t);
    showToast._t = setTimeout(() => { els.toast.hidden = true; }, 4200);
  }

  function setLoading(on) {
    els.loading.hidden = !on;
    if (on) {
      els.grid.innerHTML = "";
      els.empty.hidden = true;
    }
  }

  function rememberProfile(name, profile) {
    if (name) localStorage.setItem(ACCOUNT_KEY, name);
    if (profile) localStorage.setItem(BANCAMP_KEY, profile);
  }

  function fillSavedProfile(form) {
    const name = localStorage.getItem(ACCOUNT_KEY) || localStorage.getItem(ACCOUNT_KEY_LEGACY);
    const profile = localStorage.getItem(BANCAMP_KEY);
    const nameInput = form.querySelector('[name="account_name"]');
    const profileInput = form.querySelector('[name="bancamp_profile"]');
    if (name && nameInput && !nameInput.value) nameInput.value = name;
    if (profile && profileInput && !profileInput.value) profileInput.value = profile;
  }

  function renderAccountHTML(c) {
    const name = escapeHtml(accountNameOf(c));
    const url = c.platform_profile_url || buildProfileURL(c.platform, accountNameOf(c));
    if (url) {
      return `<a class="case-account-link" href="${escapeHtml(url)}" target="_blank" rel="noopener" onclick="event.stopPropagation()">${name} ↗</a>`;
    }
    return name;
  }

  function updateProfilePreview(platformSel, accountInput, previewEl) {
    const platform = platformSel?.value || "";
    const account = accountInput?.value?.trim() || "";
    const url = buildProfileURL(platform, account);
    if (!url) {
      previewEl.hidden = true;
      previewEl.innerHTML = "";
      return;
    }
    previewEl.hidden = false;
    previewEl.innerHTML = `Profile link: <a href="${escapeHtml(url)}" target="_blank" rel="noopener">${escapeHtml(url)}</a>`;
  }

  function populateFilters(cases) {
    const platforms = [...new Set(cases.map((c) => c.platform).filter(Boolean))].sort();
    const incidents = [...new Set(cases.map((c) => c.incident_type).filter(Boolean))].sort();

    const pf = els.filterPlatform;
    const curP = pf.value;
    pf.innerHTML = '<option value="">All platforms</option>';
    platforms.forEach((p) => {
      const o = document.createElement("option");
      o.value = p;
      o.textContent = p;
      pf.appendChild(o);
    });
    pf.value = curP;

    const fi = els.filterIncident;
    const curI = fi.value;
    fi.innerHTML = '<option value="">All incident types</option>';
    incidents.forEach((p) => {
      const o = document.createElement("option");
      o.value = p;
      o.textContent = p;
      fi.appendChild(o);
    });
    fi.value = curI;
  }

  function updateStats(cases) {
    els.statTotal.textContent = cases.length;
    const platforms = new Set(cases.map((c) => c.platform).filter(Boolean));
    els.statPlatforms.textContent = platforms.size;
    renderAnalytics(cases);
  }

  function countBy(cases, field) {
    const map = new Map();
    for (const c of cases) {
      const key = (c[field] || "").trim() || "Unknown";
      map.set(key, (map.get(key) || 0) + 1);
    }
    return [...map.entries()]
      .sort((a, b) => b[1] - a[1])
      .map(([label, count]) => ({ label, count }));
  }

  function countTemplates(cases) {
    const titles = Object.fromEntries((allTemplates || []).map((t) => [t.id, t.title]));
    const map = new Map();
    for (const c of cases) {
      if (!c.template_id) continue;
      const label = titles[c.template_id] || c.template_id;
      map.set(label, (map.get(label) || 0) + 1);
    }
    return [...map.entries()]
      .sort((a, b) => b[1] - a[1])
      .map(([label, count]) => ({ label, count }));
  }

  function countByMonth(cases) {
    const map = new Map();
    for (const c of cases) {
      if (!c.timestamp) continue;
      const d = new Date(c.timestamp);
      if (Number.isNaN(d.getTime())) continue;
      const key = `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, "0")}`;
      map.set(key, (map.get(key) || 0) + 1);
    }
    return [...map.entries()].sort((a, b) => a[0].localeCompare(b[0]));
  }

  function buildComboMatrix(cases, topN = 6) {
    const platforms = countBy(cases, "platform").slice(0, topN).map((x) => x.label);
    const incidents = countBy(cases, "incident_type").slice(0, topN).map((x) => x.label);
    const counts = new Map();
    let max = 0;

    for (const c of cases) {
      const p = (c.platform || "Unknown").trim();
      const i = (c.incident_type || "Unknown").trim();
      if (!platforms.includes(p) || !incidents.includes(i)) continue;
      const key = `${p}|||${i}`;
      const n = (counts.get(key) || 0) + 1;
      counts.set(key, n);
      if (n > max) max = n;
    }
    return { platforms, incidents, counts, max };
  }

  function chartEmpty(msg = "Not enough data yet") {
    return `<p class="chart-empty">${escapeHtml(msg)}</p>`;
  }

  function renderBarChart(container, items, fillClass, filterField) {
    if (!container) return;
    if (!items.length) {
      container.innerHTML = chartEmpty();
      return;
    }
    const max = items[0].count;
    container.innerHTML = `<div class="bar-chart">${items.map((item) => {
      const pct = max ? Math.round((item.count / max) * 100) : 0;
      const safeVal = item.label.replace(/"/g, "&quot;");
      const labelHtml = filterField
        ? `<button type="button" class="bar-label bar-label-clickable" data-filter-field="${filterField}" data-filter-value="${safeVal}" title="Filter ledger">${escapeHtml(item.label)}</button>`
        : `<span class="bar-label" title="${escapeHtml(item.label)}">${escapeHtml(item.label)}</span>`;
      return `
        <div class="bar-row">
          ${labelHtml}
          <div class="bar-track"><div class="bar-fill ${fillClass}" style="width:${pct}%"></div></div>
          <span class="bar-value">${item.count}</span>
        </div>`;
    }).join("")}</div>`;

    if (filterField) {
      container.querySelectorAll("[data-filter-field]").forEach((btn) => {
        btn.addEventListener("click", () => applyLedgerFilter(btn.dataset.filterField, btn.dataset.filterValue));
      });
    }
  }

  function applyLedgerFilter(field, value) {
    if (field === "platform") {
      els.filterPlatform.value = value;
      els.filterIncident.value = "";
    } else if (field === "incident_type") {
      els.filterIncident.value = value;
      els.filterPlatform.value = "";
    }
    els.search.value = "";
    renderGrid();
    document.getElementById("ledger").scrollIntoView({ behavior: "smooth" });
  }

  function renderMatrix(container, matrix) {
    if (!container) return;
    const { platforms, incidents, counts, max } = matrix;
    if (!platforms.length || !incidents.length) {
      container.innerHTML = chartEmpty();
      return;
    }

    const header = `<tr><th class="matrix-corner">Platform ↓ / Issue →</th>${incidents.map((i) => `<th>${escapeHtml(truncate(i, 18))}</th>`).join("")}</tr>`;
    const rows = platforms.map((p) => {
      const cells = incidents.map((i) => {
        const n = counts.get(`${p}|||${i}`) || 0;
        if (!n) return `<td class="matrix-cell matrix-cell-empty">·</td>`;
        const intensity = max ? n / max : 0;
        const bg = `rgba(255, 77, 77, ${0.15 + intensity * 0.75})`;
        const hot = intensity > 0.55 ? " matrix-cell-hot" : "";
        return `<td class="matrix-cell${hot}" style="background:${bg}">${n}</td>`;
      }).join("");
      return `<tr><td class="matrix-rowhead">${escapeHtml(truncate(p, 20))}</td>${cells}</tr>`;
    }).join("");

    container.innerHTML = `<div class="matrix-wrap"><table class="matrix-table"><thead>${header}</thead><tbody>${rows}</tbody></table></div>`;
  }

  function renderTimeline(container, months) {
    if (!container) return;
    if (!months.length) {
      container.innerHTML = chartEmpty();
      return;
    }

    const recent = months.slice(-12);
    const max = Math.max(...recent.map((m) => m.count), 1);

    container.innerHTML = `<div class="timeline-chart">${recent.map((m) => {
      const h = Math.max(8, Math.round((m.count / max) * 100));
      const [y, mo] = m.label.split("-");
      const short = new Date(Date.UTC(Number(y), Number(mo) - 1)).toLocaleString("en", { month: "short" });
      return `
        <div class="timeline-col" title="${m.label}: ${m.count} cases">
          <span class="timeline-count">${m.count}</span>
          <div class="timeline-bar-wrap"><div class="timeline-bar" style="height:${h}%"></div></div>
          <span class="timeline-label">${short} '${String(y).slice(2)}</span>
        </div>`;
    }).join("")}</div>`;
  }

  function renderOverview(container, cases) {
    if (!container) return;
    if (!cases.length) {
      container.innerHTML = "";
      return;
    }

    const platforms = countBy(cases, "platform");
    const incidents = countBy(cases, "incident_type");
    const signals = cases.filter((c) => c.submission_type === "signal").length;
    const reports = cases.length - signals;
    const verified = cases.filter((c) => c.verified).length;
    const withProof = cases.filter((c) => c.proof_file_name).length;
    const edited = cases.filter((c) => c.edited_at).length;
    const topCombo = (() => {
      const m = new Map();
      for (const c of cases) {
        const k = `${c.platform || "?"} + ${c.incident_type || "?"}`;
        m.set(k, (m.get(k) || 0) + 1);
      }
      return [...m.entries()].sort((a, b) => b[1] - a[1])[0];
    })();

    const cards = [
      { value: platforms[0] ? `${platforms[0].label} (${platforms[0].count})` : "—", label: "Worst platform", cls: "overview-value-warn" },
      { value: incidents[0] ? truncate(incidents[0].label, 22) : "—", label: `Top issue (${incidents[0]?.count ?? 0})`, cls: "overview-value-warn" },
      { value: `${signals} / ${reports}`, label: "Signals vs full reports", cls: "overview-value-accent" },
      { value: `${verified}`, label: `Verified cases (${cases.length ? Math.round((verified / cases.length) * 100) : 0}%)`, cls: "overview-value-trust" },
      { value: `${withProof}`, label: "Entries with proof", cls: "" },
      { value: topCombo ? truncate(topCombo[0], 28) : "—", label: `Top combo (${topCombo?.[1] ?? 0})`, cls: "overview-value-warn" },
      { value: `${edited}`, label: "Edited entries", cls: "" },
    ];

    container.innerHTML = cards.map((c) => `
      <div class="overview-card">
        <span class="overview-value ${c.cls}">${escapeHtml(String(c.value))}</span>
        <span class="overview-label">${escapeHtml(c.label)}</span>
      </div>`).join("");
  }

  function renderAnalytics(cases) {
    renderOverview($("#analytics-overview"), cases);
    renderBarChart($("#chart-platforms"), countBy(cases, "platform").slice(0, 8), "bar-fill-platform", "platform");
    renderBarChart($("#chart-incidents"), countBy(cases, "incident_type").slice(0, 8), "bar-fill-incident", "incident_type");
    renderMatrix($("#chart-matrix"), buildComboMatrix(cases, 6));
    renderBarChart($("#chart-templates"), countTemplates(cases).slice(0, 8), "bar-fill-template");
    renderBarChart($("#chart-reasons"), countBy(cases, "reason_category").filter((x) => x.label !== "Unknown").slice(0, 8), "bar-fill-reason");
    renderTimeline($("#chart-timeline"), countByMonth(cases).map(([label, count]) => ({ label, count })));
  }

  function renderTemplates() {
    if (!allTemplates.length) {
      els.templatesGrid.innerHTML = '<p class="templates-loading">Loading common cases…</p>';
      return;
    }

    els.templatesGrid.innerHTML = allTemplates.map((t, i) => `
      <button type="button" class="template-card" data-template-id="${escapeHtml(t.id)}" style="animation-delay:${i * 0.04}s">
        <span class="template-count">${t.count > 0 ? t.count : "—"}</span>
        <span class="template-title">${escapeHtml(t.title)}</span>
        <span class="template-desc">${escapeHtml(t.description)}</span>
        <span class="template-cta">Same issue here →</span>
      </button>
    `).join("");

    $$(".template-card").forEach((btn) => {
      btn.addEventListener("click", () => {
        const tmpl = allTemplates.find((x) => x.id === btn.dataset.templateId);
        if (tmpl) openSignalModal(tmpl);
      });
    });
  }

  function filterCases() {
    const q = els.search.value.trim().toLowerCase();
    const platform = els.filterPlatform.value;
    const incident = els.filterIncident.value;
    const sort = els.sortBy.value;

    let list = allCases.filter((c) => {
      if (platform && c.platform !== platform) return false;
      if (incident && c.incident_type !== incident) return false;
      if (!q) return true;
      const hay = [
        accountNameOf(c),
        c.platform,
        c.incident_type,
        c.reason_category,
        c.story,
        c.submission_type,
      ].join(" ").toLowerCase();
      return hay.includes(q);
    });

    list = [...list];
    if (sort === "oldest") {
      list.sort((a, b) => a.timestamp.localeCompare(b.timestamp));
    } else if (sort === "account") {
      list.sort((a, b) => accountNameOf(a).localeCompare(accountNameOf(b)));
    } else {
      list.sort((a, b) => b.timestamp.localeCompare(a.timestamp));
    }

    return list;
  }

  function storyPreview(c) {
    if (c.story) return truncate(c.story, 180);
    if (c.submission_type === "signal") return "Quick signal — no additional story provided.";
    return "No story provided.";
  }

  function renderCard(c, index) {
    const id = String(c.id).padStart(4, "0");
    const verified = c.verified ? '<span class="case-verified">Verified</span>' : "";
    const typeClass = c.submission_type === "signal" ? "case-type-signal" : "case-type-report";
    return `
      <article class="case-card" role="listitem" data-id="${c.id}" style="animation-delay:${index * 0.05}s">
        <div class="case-card-header">
          <span class="case-id">ENTRY_${id}</span>
          <span class="case-type ${typeClass}">${submissionLabel(c)}</span>
          ${verified}
        </div>
        <h3 class="case-artist">${renderAccountHTML(c)}</h3>
        <div class="case-platform">${escapeHtml(c.platform)}</div>
        <div class="case-incident">${escapeHtml(c.incident_type)}</div>
        <p class="case-story-preview">${escapeHtml(storyPreview(c))}</p>
        <div class="case-footer">
          <span>${formatDate(c.timestamp)}</span>
          <span>View →</span>
        </div>
      </article>
    `;
  }

  function renderGrid() {
    const list = filterCases();
    els.grid.innerHTML = list.map((c, i) => renderCard(c, i)).join("");
    els.empty.hidden = list.length > 0 || allCases.length > 0;
    if (allCases.length === 0) els.empty.hidden = false;

    $$(".case-card").forEach((card) => {
      card.addEventListener("click", () => {
        const id = Number(card.dataset.id);
        const c = allCases.find((x) => x.id === id);
        if (c) openCaseModal(c);
      });
    });
  }

  function renderProof(c) {
    const fn = c.proof_file_name;
    if (!fn) return "";
    const url = proofUrl(fn);
    const ext = getExt(fn);

    if (isImage(ext)) {
      return `<img src="${url}" alt="Proof for entry ${c.id}" loading="lazy">`;
    }
    if (ext === ".pdf") {
      return `
        <iframe src="${url}" title="Proof PDF" width="100%" height="480" style="border:1px solid var(--border);border-radius:12px;"></iframe>
        <a class="proof-link" href="${url}" target="_blank" rel="noopener">Open PDF in new tab ↗</a>
      `;
    }
    return `<a class="proof-link" href="${url}" target="_blank" rel="noopener">Download proof file (${escapeHtml(fn)}) ↗</a>`;
  }

  function openCaseModal(c) {
    activeCase = c;
    const id = String(c.id).padStart(4, "0");
    $("#modal-id").textContent = `ENTRY_${id}`;
    const profileURL = c.platform_profile_url || buildProfileURL(c.platform, accountNameOf(c));
    $("#modal-title").innerHTML = profileURL
      ? `<a href="${escapeHtml(profileURL)}" target="_blank" rel="noopener" class="case-account-link">${escapeHtml(accountNameOf(c))} ↗</a>`
      : escapeHtml(accountNameOf(c));

    let meta = `Logged ${formatDate(c.timestamp)} · ${submissionLabel(c)} · ${escapeHtml(c.platform)}`;
    if (c.edited_at) {
      meta += ` · edited ${formatDate(c.edited_at)}`;
    }
    if (c.bancamp_profile) {
      meta += ` · <a href="${escapeHtml(c.bancamp_profile)}" target="_blank" rel="noopener">Bancamp ↗</a>`;
    }
    $("#modal-meta").innerHTML = meta;

    els.modalFooter.hidden = !editEnabled;

    const tags = [
      `<span class="tag">${escapeHtml(c.platform)}</span>`,
      `<span class="tag">${escapeHtml(c.incident_type)}</span>`,
    ];
    if (c.reason_category) {
      tags.push(`<span class="tag tag-reason">${escapeHtml(c.reason_category)}</span>`);
    }
    if (c.verified) {
      tags.push('<span class="tag" style="background:rgba(110,231,168,0.12);color:var(--verified);border-color:rgba(110,231,168,0.3)">Verified</span>');
    }
    $("#modal-tags").innerHTML = tags.join("");

    const storyEl = $("#modal-story");
    if (c.story) {
      storyEl.textContent = c.story;
      storyEl.hidden = false;
    } else {
      storyEl.textContent = "";
      storyEl.hidden = true;
    }

    const proofHtml = renderProof(c);
    const proofEl = $("#modal-proof");
    if (proofHtml) {
      proofEl.innerHTML = proofHtml;
      proofEl.hidden = false;
    } else {
      proofEl.innerHTML = "";
      proofEl.hidden = true;
    }

    els.caseModal.showModal();
  }

  function openEditModal() {
    if (!activeCase) return;
    const c = activeCase;
    els.editError.hidden = true;
    els.editForm.reset();

    $("#edit_entry_label").textContent = `ENTRY_${String(c.id).padStart(4, "0")}`;
    $("#edit_entry_id").value = c.id;
    $("#edit_account_name").value = accountNameOf(c);
    $("#edit_platform").value = c.platform || "Other";
    $("#edit_bancamp_profile").value = c.bancamp_profile || "";
    $("#edit_incident_type").value = c.incident_type || "";
    $("#edit_reason_category").value = c.reason_category || "";
    $("#edit_story").value = c.story || "";
    $("#edit_verified").checked = !!c.verified;
    $("#edit_remove_proof").checked = false;

    const savedToken = sessionStorage.getItem(EDIT_TOKEN_KEY);
    if (savedToken) $("#edit_token").value = savedToken;

    resetFileDrop(els.editFileDrop, els.editFileDropText, els.editProofInput, "Upload new screenshot or PDF — max 5MB");
    updateProfilePreview($("#edit_platform"), $("#edit_account_name"), $("#edit-profile-preview"));

    els.caseModal.close();
    els.editModal.showModal();
  }

  async function loadMeta() {
    try {
      const res = await fetch("/api/meta");
      if (!res.ok) return;
      const data = await res.json();
      editEnabled = !!data.edit_enabled;
    } catch {
      editEnabled = false;
    }
  }

  async function saveEdit(e) {
    e.preventDefault();
    els.editError.hidden = true;

    const token = $("#edit_token").value.trim();
    if (!token) {
      els.editError.textContent = "Edit token is required";
      els.editError.hidden = false;
      return;
    }

    const btn = els.editBtn;
    const label = btn.querySelector(".btn-label");
    const spinner = btn.querySelector(".btn-spinner");
    btn.disabled = true;
    label.hidden = true;
    spinner.hidden = false;

    const formData = new FormData(els.editForm);
    formData.set("verified", $("#edit_verified").checked ? "true" : "false");
    if (!$("#edit_remove_proof").checked) {
      formData.delete("remove_proof");
    }

    const entryId = activeCase?.id;
    try {
      const res = await fetch(`/api/cases/${entryId}`, {
        method: "PATCH",
        headers: { "X-Edit-Token": token },
        body: formData,
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || `Edit failed (${res.status})`);

      sessionStorage.setItem(EDIT_TOKEN_KEY, token);
      els.editModal.close();
      showToast(data.message || "Entry updated.");
      await loadCases();
    } catch (err) {
      els.editError.textContent = err.message;
      els.editError.hidden = false;
    } finally {
      btn.disabled = false;
      label.hidden = false;
      spinner.hidden = true;
    }
  }

  function openSignalModal(tmpl) {
    activeTemplate = tmpl;
    els.signalError.hidden = true;
    els.signalForm.reset();
    $("#signal_template_id").value = tmpl.id;
    $("#signal-template-label").textContent = "COMMON CASE";
    $("#signal-title").textContent = tmpl.title;
    $("#signal-desc").textContent = tmpl.description;
    resetFileDrop(els.signalFileDrop, els.signalFileDropText, els.signalProofInput, "Screenshot or PDF — max 5MB");
    fillSavedProfile(els.signalForm);
    updateProfilePreview($("#signal_platform"), $("#signal_account_name"), $("#signal-profile-preview"));
    els.signalModal.showModal();
  }

  function openSubmitModal(e) {
    if (e) e.preventDefault();
    els.formError.hidden = true;
    els.submitForm.reset();
    resetFileDrop(els.fileDrop, els.fileDropText, els.proofInput, "Drop screenshot, PDF, or email (.eml) — max 5MB");
    fillSavedProfile(els.submitForm);
    updateProfilePreview($("#platform"), $("#account_name"), $("#report-profile-preview"));
    els.submitModal.showModal();
  }

  function resetFileDrop(drop, textEl, input, placeholder) {
    drop.classList.remove("has-file");
    textEl.textContent = placeholder;
    if (input) input.value = "";
  }

  async function loadTemplates() {
    try {
      const res = await fetch("/api/templates");
      if (!res.ok) throw new Error("templates failed");
      allTemplates = await res.json();
      renderTemplates();
    } catch {
      els.templatesGrid.innerHTML = '<p class="templates-loading">Could not load case templates.</p>';
    }
  }

  async function loadCases() {
    setLoading(true);
    try {
      const res = await fetch("/api/cases");
      if (!res.ok) throw new Error("Failed to load cases");
      allCases = await res.json();
      updateStats(allCases);
      populateFilters(allCases);
      renderGrid();
      await loadTemplates();
      renderAnalytics(allCases);
    } catch {
      showToast("Could not load cases. Is the server running?", "error");
      els.empty.hidden = false;
    } finally {
      setLoading(false);
    }
  }

  async function postForm(url, form, errorEl, modal, btn) {
    errorEl.hidden = true;
    const label = btn.querySelector(".btn-label");
    const spinner = btn.querySelector(".btn-spinner");

    btn.disabled = true;
    label.hidden = true;
    spinner.hidden = false;

    const formData = new FormData(form);
    const accountName = formData.get("account_name");
    const bancampProfile = formData.get("bancamp_profile");

    try {
      const res = await fetch(url, { method: "POST", body: formData });
      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        throw new Error(data.error || `Submit failed (${res.status})`);
      }

      rememberProfile(accountName, bancampProfile);
      modal.close();
      showToast(data.message || "Recorded on the ledger.");
      await loadCases();
    } catch (err) {
      errorEl.textContent = err.message;
      errorEl.hidden = false;
    } finally {
      btn.disabled = false;
      label.hidden = false;
      spinner.hidden = true;
    }
  }

  function setupFileDrop(drop, textEl, input) {
    input.addEventListener("change", () => {
      if (input.files?.[0]) {
        drop.classList.add("has-file");
        textEl.textContent = input.files[0].name;
      }
    });

    ["dragenter", "dragover"].forEach((ev) => {
      drop.addEventListener(ev, (e) => {
        e.preventDefault();
        drop.classList.add("dragover");
      });
    });
    ["dragleave", "drop"].forEach((ev) => {
      drop.addEventListener(ev, (e) => {
        e.preventDefault();
        drop.classList.remove("dragover");
      });
    });
    drop.addEventListener("drop", (e) => {
      const files = e.dataTransfer?.files;
      if (files?.[0]) {
        input.files = files;
        drop.classList.add("has-file");
        textEl.textContent = files[0].name;
      }
    });
  }

  function bindEvents() {
    $("#open-submit").addEventListener("click", openSubmitModal);
    $("#open-submit-inline").addEventListener("click", openSubmitModal);
    $("#empty-submit").addEventListener("click", () => {
      document.getElementById("signal").scrollIntoView({ behavior: "smooth" });
    });

    $("#close-modal").addEventListener("click", () => els.caseModal.close());
    $("#edit-case-btn")?.addEventListener("click", openEditModal);
    $("#close-edit")?.addEventListener("click", () => els.editModal.close());
    $("#cancel-edit")?.addEventListener("click", () => els.editModal.close());
    $("#close-signal").addEventListener("click", () => els.signalModal.close());
    $("#close-submit").addEventListener("click", () => els.submitModal.close());
    $("#cancel-signal").addEventListener("click", () => els.signalModal.close());
    $("#cancel-submit").addEventListener("click", () => els.submitModal.close());

    [els.caseModal, els.signalModal, els.submitModal, els.editModal].forEach((modal) => {
      if (!modal) return;
      modal.addEventListener("click", (e) => {
        if (e.target === modal) modal.close();
      });
    });

    const signalPlatform = $("#signal_platform");
    const signalAccount = $("#signal_account_name");
    const signalPreview = $("#signal-profile-preview");
    const reportPlatform = $("#platform");
    const reportAccount = $("#account_name");
    const reportPreview = $("#report-profile-preview");

    const refreshSignalPreview = () => updateProfilePreview(signalPlatform, signalAccount, signalPreview);
    const refreshReportPreview = () => updateProfilePreview(reportPlatform, reportAccount, reportPreview);

    signalPlatform?.addEventListener("change", refreshSignalPreview);
    signalAccount?.addEventListener("input", refreshSignalPreview);
    reportPlatform?.addEventListener("input", refreshReportPreview);
    reportPlatform?.addEventListener("change", refreshReportPreview);
    reportAccount?.addEventListener("input", refreshReportPreview);

    const editPlatform = $("#edit_platform");
    const editAccount = $("#edit_account_name");
    const editPreview = $("#edit-profile-preview");
    const refreshEditPreview = () => updateProfilePreview(editPlatform, editAccount, editPreview);
    editPlatform?.addEventListener("change", refreshEditPreview);
    editAccount?.addEventListener("input", refreshEditPreview);

    els.editForm?.addEventListener("submit", saveEdit);

    els.signalForm.addEventListener("submit", (e) => {
      e.preventDefault();
      postForm("/api/signal", els.signalForm, els.signalError, els.signalModal, els.signalBtn);
    });

    els.submitForm.addEventListener("submit", (e) => {
      e.preventDefault();
      postForm("/api/submit-case", els.submitForm, els.formError, els.submitModal, els.submitBtn);
    });

    [els.search, els.filterPlatform, els.filterIncident, els.sortBy].forEach((el) => {
      el.addEventListener("input", renderGrid);
      el.addEventListener("change", renderGrid);
    });
  }

  setupFileDrop(els.fileDrop, els.fileDropText, els.proofInput);
  setupFileDrop(els.signalFileDrop, els.signalFileDropText, els.signalProofInput);
  setupFileDrop(els.editFileDrop, els.editFileDropText, els.editProofInput);
  bindEvents();
  loadMeta().then(loadCases);
})();