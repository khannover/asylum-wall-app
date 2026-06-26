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
    const id = String(c.id).padStart(4, "0");
    $("#modal-id").textContent = `ENTRY_${id}`;
    const profileURL = c.platform_profile_url || buildProfileURL(c.platform, accountNameOf(c));
    $("#modal-title").innerHTML = profileURL
      ? `<a href="${escapeHtml(profileURL)}" target="_blank" rel="noopener" class="case-account-link">${escapeHtml(accountNameOf(c))} ↗</a>`
      : escapeHtml(accountNameOf(c));

    let meta = `${formatDate(c.timestamp)} · ${submissionLabel(c)} · ${escapeHtml(c.platform)}`;
    if (c.bancamp_profile) {
      meta += ` · <a href="${escapeHtml(c.bancamp_profile)}" target="_blank" rel="noopener">Bancamp ↗</a>`;
    }
    $("#modal-meta").innerHTML = meta;

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
    $("#close-signal").addEventListener("click", () => els.signalModal.close());
    $("#close-submit").addEventListener("click", () => els.submitModal.close());
    $("#cancel-signal").addEventListener("click", () => els.signalModal.close());
    $("#cancel-submit").addEventListener("click", () => els.submitModal.close());

    [els.caseModal, els.signalModal, els.submitModal].forEach((modal) => {
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
  bindEvents();
  loadCases();
})();