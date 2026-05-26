const API_BASE = "/api/swipe";

let profiles = [];
let currentIndex = 0;
let isDragging = false;
let dragStartX = 0;
let dragStartY = 0;
let currentCard = null;
let currentCardStartX = 0;
let currentCardStartY = 0;

function qs(sel) {
  return document.querySelector(sel);
}

function qsa(sel) {
  return Array.from(document.querySelectorAll(sel));
}

async function fetchProfiles() {
  const loading = qs("#loading-state");
  const empty = qs("#empty-state");
  const actions = qs("#swipe-actions");
  const stack = qs("#card-stack");

  loading.style.display = "flex";
  empty.classList.remove("visible");
  actions.style.display = "none";

  // Remove old cards
  qsa(".swipe-card").forEach((el) => el.remove());

  try {
    const res = await fetch(`${API_BASE}/profiles`, {
      headers: { Accept: "application/json" },
    });
    if (!res.ok) {
      throw new Error("fetch error");
    }
    profiles = await res.json();
  } catch (e) {
    console.error(e);
    profiles = [];
  }

  loading.style.display = "none";

  if (!profiles || profiles.length === 0) {
    empty.classList.add("visible");
    actions.style.display = "none";
    currentIndex = 0;
    currentCard = null;
    return;
  }

  empty.classList.remove("visible");
  actions.style.display = "flex";

  currentIndex = 0;
  buildCards(stack, profiles);
  setTopCard();
}

function buildCards(stack, data) {
  data.forEach((profile, idx) => {
    const card = document.createElement("div");
    card.className = "swipe-card";
    card.dataset.index = idx.toString();
    card.style.zIndex = (1000 - idx).toString();

    const stampLike = document.createElement("div");
    stampLike.className = "stamp stamp-like";
    stampLike.textContent = "LIKE";

    const stampPass = document.createElement("div");
    stampPass.className = "stamp stamp-pass";
    stampPass.textContent = "PASS";

    const photosWrap = document.createElement("div");
    photosWrap.className = "card-photos";

    const hasPhotos = profile.photos && profile.photos.length > 0;
    if (hasPhotos) {
      profile.photos.forEach((p, i) => {
        const ph = document.createElement("div");
        ph.className = "card-photo";
        if (i === 0) ph.classList.add("active");
        ph.style.backgroundImage = `url(${p})`;
        photosWrap.appendChild(ph);
      });
    } else {
      const def = document.createElement("div");
      def.className = "card-photo-default";
      const span = document.createElement("span");
      span.className = "avatar-text";
      span.textContent =
        (profile.username && profile.username[0]) ||
        (profile.first_name && profile.first_name[0]) ||
        "?";
      def.appendChild(span);
      photosWrap.appendChild(def);
    }

    const photoGradient = document.createElement("div");
    photoGradient.className = "photo-gradient";
    photosWrap.appendChild(photoGradient);

    const body = document.createElement("div");
    body.className = "card-body";

    const nameRow = document.createElement("div");
    nameRow.className = "card-name";
    const nameSpan = document.createElement("span");
    nameSpan.textContent =
      (profile.first_name || "") + " " + (profile.last_name || "");
    const usernameSpan = document.createElement("span");
    usernameSpan.className = "card-username";
    usernameSpan.textContent = "@" + profile.username;
    nameRow.appendChild(nameSpan);
    nameRow.appendChild(usernameSpan);

    const chipsRow = document.createElement("div");
    chipsRow.className = "card-game-types";
    if (profile.game_types && profile.game_types.length > 0) {
      profile.game_types.forEach((gt) => {
        const chip = document.createElement("div");
        chip.className = "game-type-chip chip-other";
        chip.textContent = gt;
        chipsRow.appendChild(chip);
      });
    }

    const divider = document.createElement("div");
    divider.className = "card-divider";

    const bioTitle = document.createElement("div");
    bioTitle.className = "card-section-title";
    bioTitle.textContent = "Bio";

    const bio = document.createElement("p");
    bio.className = "card-bio";
    bio.textContent = profile.biography || "Ce joueur n'a pas encore écrit de bio.";

    const metaRow = document.createElement("div");
    metaRow.className = "card-meta-row";

    if (profile.play_style) {
      const tag = document.createElement("div");
      tag.className = "card-meta-tag";
      tag.innerHTML = `Style&nbsp;: <span>${profile.play_style}</span>`;
      metaRow.appendChild(tag);
    }

    if (profile.skill_level) {
      const tag = document.createElement("div");
      tag.className = "card-meta-tag";
      tag.innerHTML = `Niveau&nbsp;: <span>${profile.skill_level}</span>`;
      metaRow.appendChild(tag);
    }

    const gamesRow = document.createElement("div");
    gamesRow.className = "card-games";
    if (profile.games_played && profile.games_played.length > 0) {
      profile.games_played.forEach((g) => {
        const pill = document.createElement("div");
        pill.className = "card-game-pill";
        pill.textContent = g;
        gamesRow.appendChild(pill);
      });
    }

    body.appendChild(nameRow);
    body.appendChild(chipsRow);
    body.appendChild(divider);
    body.appendChild(bioTitle);
    body.appendChild(bio);
    body.appendChild(metaRow);
    body.appendChild(gamesRow);

    card.appendChild(stampLike);
    card.appendChild(stampPass);
    card.appendChild(photosWrap);
    card.appendChild(body);

    stack.appendChild(card);
  });
}

function setTopCard() {
  const cards = qsa(".swipe-card");
  cards.forEach((c) => c.classList.remove("is-top"));

  if (cards.length === 0) {
    currentCard = null;
    return;
  }

  // top visuel = dernière carte dans le DOM
  currentCard = cards[cards.length - 1];
  currentCard.classList.add("is-top");

  resetCardTransform(currentCard);
}

function resetCardTransform(card) {
  card.style.transform = "";
  card.style.transition = "";
  card.classList.remove("is-dragging", "hint-like", "hint-pass");
}

function attachEvents() {
  const passBtn = qs("#btn-pass");
  const likeBtn = qs("#btn-like");
  const reloadBtn = qs("#btn-reload-profiles");

  if (passBtn) {
    passBtn.addEventListener("click", () => {
      if (!currentCard) return;
      handleDecision("pass");
    });
  }

  if (likeBtn) {
    likeBtn.addEventListener("click", () => {
      if (!currentCard) return;
      handleDecision("like");
    });
  }

  if (reloadBtn) {
    reloadBtn.addEventListener("click", () => {
      fetchProfiles();
    });
  }

  const stack = qs("#card-stack");
  stack.addEventListener("pointerdown", onPointerDown);
  window.addEventListener("pointermove", onPointerMove);
  window.addEventListener("pointerup", onPointerUp);
}

function onPointerDown(e) {
  if (!currentCard) return;
  if (e.button !== 0 && e.pointerType === "mouse") return;

  isDragging = true;
  dragStartX = e.clientX;
  dragStartY = e.clientY;

  const rect = currentCard.getBoundingClientRect();
  currentCardStartX = rect.left + rect.width / 2;
  currentCardStartY = rect.top + rect.height / 2;

  currentCard.classList.add("is-dragging");
  currentCard.style.transition = "none";
}

function onPointerMove(e) {
  if (!isDragging || !currentCard) return;

  const dx = e.clientX - dragStartX;
  const dy = e.clientY - dragStartY;

  const rotation = dx / 10;
  const threshold = 80;

  currentCard.style.transform = `translate(${dx}px, ${dy}px) rotate(${rotation}deg)`;

  currentCard.classList.remove("hint-like", "hint-pass");
  if (dx > threshold) {
    currentCard.classList.add("hint-like");
  } else if (dx < -threshold) {
    currentCard.classList.add("hint-pass");
  }
}

function onPointerUp(e) {
  if (!isDragging || !currentCard) return;
  isDragging = false;

  const rect = currentCard.getBoundingClientRect();
  const dx = rect.left + rect.width / 2 - currentCardStartX;
  const threshold = 80;

  currentCard.style.transition = "transform .2s ease-out";

  if (dx > threshold) {
    currentCard.classList.add("fly-right");
    const idx = Number(currentCard.dataset.index);
    sendSwipe(profiles[idx].id, "like");
    removeCurrentCardAfterAnim();
  } else if (dx < -threshold) {
    currentCard.classList.add("fly-left");
    const idx = Number(currentCard.dataset.index);
    sendSwipe(profiles[idx].id, "pass");
    removeCurrentCardAfterAnim();
  } else {
    currentCard.style.transform = "";
    currentCard.classList.remove("hint-like", "hint-pass", "is-dragging");
  }
}

function removeCurrentCardAfterAnim() {
  if (!currentCard) return;
  const card = currentCard;
  setTimeout(() => {
    card.remove();
    setTopCardIfAny();
  }, 200);
}

function setTopCardIfAny() {
  const cards = qsa(".swipe-card");
  if (cards.length === 0) {
    currentCard = null;
    const empty = qs("#empty-state");
    const actions = qs("#swipe-actions");
    empty.classList.add("visible");
    actions.style.display = "none";
    return;
  }
  setTopCard();
}

async function sendSwipe(targetId, direction) {
  try {
    const res = await fetch(`${API_BASE}/action`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ target_id: targetId, direction }),
    });
    if (!res.ok) return;
    const data = await res.json();
    if (data.matched) {
      openMatchModal(targetId);
      fetchMatches(); // on rafraîchit la liste des matchs
    }
  } catch (e) {
    console.error(e);
  }
}

function handleDecision(direction) {
  if (!currentCard) return;
  currentCard.style.transition = "transform .2s ease-out";
  const idx = Number(currentCard.dataset.index);
  if (direction === "like") {
    currentCard.classList.add("fly-right");
    sendSwipe(profiles[idx].id, "like");
  } else {
    currentCard.classList.add("fly-left");
    sendSwipe(profiles[idx].id, "pass");
  }
  removeCurrentCardAfterAnim();
}

function openMatchModal(targetId) {
  const modal = qs("#match-modal");
  const themText = qs("#match-them-text");
  const meText = qs("#match-me-text");

  const p = profiles.find((u) => u.id === targetId);
  if (p) {
    themText.textContent =
      (p.username && p.username[0].toUpperCase()) || "?";
  }
  meText.textContent = "U";

  modal.classList.add("visible");

  const btnContinue = qs("#btn-continue-swipe");
  const btnViewMatches = qs("#btn-view-matches");

  btnContinue.onclick = () => {
    modal.classList.remove("visible");
  };
  btnViewMatches.onclick = () => {
    modal.classList.remove("visible");
    document.getElementById("view-swipe").style.display = "none";
    const viewMatches = document.getElementById("view-matches");
    viewMatches.classList.add("visible");
  };
}

/* ---------- MATCHES (sidebar + full view) ---------- */

async function fetchMatches() {
  try {
    const res = await fetch(`${API_BASE}/matches`, {
      headers: { Accept: "application/json" },
    });
    if (!res.ok) throw new Error("matches fetch error");

    const matches = await res.json();
    renderMatchesSidebar(matches);
    renderMatchesGrid(matches);
  } catch (e) {
    console.error("Error fetching matches:", e);
  }
}

function renderMatchesSidebar(matches) {
  const list = qs("#matches-sidebar-list");
  const countBadge = qs("#match-sidebar-count");

  list.innerHTML = "";

  if (!matches || matches.length === 0) {
    const empty = document.createElement("div");
    empty.className = "matches-empty";
    empty.innerHTML = `
      <div class="matches-empty-icon">🎯</div>
      <p>Tes smash apparaîtront ici. Continue à swiper !</p>
    `;
    list.appendChild(empty);
    countBadge.textContent = "0";
    return;
  }

  countBadge.textContent = matches.length.toString();

  matches.forEach((m) => {
    const u = m.user || {};
    const card = document.createElement("div");
    card.className = "match-card";

    const avatar = document.createElement("div");
    avatar.className = "match-avatar";

    if (u.photos && u.photos.length > 0) {
      const img = document.createElement("img");
      img.src = u.photos[0];
      avatar.appendChild(img);
    }

    const txt = document.createElement("div");
    txt.className = "match-avatar-text";
    txt.textContent = (u.username && u.username[0].toUpperCase()) || "?";
    avatar.appendChild(txt);

    const info = document.createElement("div");
    info.className = "match-info";

    const name = document.createElement("div");
    name.className = "match-name";
    name.textContent =
      (u.first_name || "") + " " + (u.last_name || "");

    const tags = document.createElement("div");
    tags.className = "match-tags";

    if (u.play_style) {
      const t = document.createElement("div");
      t.className = "match-tag";
      t.textContent = u.play_style;
      tags.appendChild(t);
    }

    if (u.skill_level) {
      const t = document.createElement("div");
      t.className = "match-tag";
      t.textContent = u.skill_level;
      tags.appendChild(t);
    }

    info.appendChild(name);
    info.appendChild(tags);

    const dot = document.createElement("div");
    dot.className = "match-new";

    card.appendChild(avatar);
    card.appendChild(info);
    card.appendChild(dot);

    list.appendChild(card);
  });
}

function renderMatchesGrid(matches) {
  const grid = qs("#matches-full-grid");
  grid.innerHTML = "";

  if (!matches || matches.length === 0) {
    const empty = document.createElement("div");
    empty.className = "matches-empty";
    empty.style.gridColumn = "1/-1";
    empty.style.padding = "4rem 0";
    empty.innerHTML = `
      <div class="matches-empty-icon">🎮</div>
      <p>Aucun smash pour le moment. Va swiper !</p>
    `;
    grid.appendChild(empty);
    return;
  }

  matches.forEach((m) => {
    const u = m.user || {};
    const card = document.createElement("div");
    card.className = "match-full-card";

    const top = document.createElement("div");
    top.className = "match-full-top";

    const avatar = document.createElement("div");
    avatar.className = "match-full-avatar";

    if (u.photos && u.photos.length > 0) {
      const img = document.createElement("img");
      img.src = u.photos[0];
      avatar.appendChild(img);
    } else {
      const span = document.createElement("span");
      span.className = "match-m-avatar-text";
      span.textContent =
        (u.username && u.username[0].toUpperCase()) || "?";
      avatar.appendChild(span);
    }

    const info = document.createElement("div");
    const name = document.createElement("div");
    name.className = "match-full-name";
    name.textContent =
      (u.first_name || "") + " " + (u.last_name || "");

    const username = document.createElement("div");
    username.className = "match-full-username";
    username.textContent = "@" + (u.username || "unknown");

    const date = document.createElement("div");
    date.className = "match-full-date";
    date.textContent = formatMatchDate(m.matched_at);

    info.appendChild(name);
    info.appendChild(username);
    info.appendChild(date);

    top.appendChild(avatar);
    top.appendChild(info);

    const tags = document.createElement("div");
    tags.className = "match-full-tags";

    if (u.play_style) {
      const t = document.createElement("div");
      t.className = "match-tag";
      t.textContent = u.play_style;
      tags.appendChild(t);
    }

    if (u.skill_level) {
      const t = document.createElement("div");
      t.className = "match-tag";
      t.textContent = u.skill_level;
      tags.appendChild(t);
    }

    const actions = document.createElement("div");
    actions.className = "match-full-actions";

    const btnProfile = document.createElement("button");
    btnProfile.className = "btn btn-primary btn-sm";
    btnProfile.textContent = "Voir le profil";
    btnProfile.addEventListener("click", () => {
      window.location.href = `/users/${u.id}`;
    });

    const btnUnmatch = document.createElement("button");
    btnUnmatch.className = "btn btn-unmatch btn-sm";
    btnUnmatch.textContent = "Unmatch";
    btnUnmatch.addEventListener("click", () => {
      handleUnmatch(m.match_id);
    });

    actions.appendChild(btnProfile);
    actions.appendChild(btnUnmatch);

    card.appendChild(top);
    card.appendChild(tags);
    card.appendChild(actions);

    grid.appendChild(card);
  });
}

function formatMatchDate(isoString) {
  if (!isoString) return "";
  const d = new Date(isoString);
  if (isNaN(d.getTime())) return "";
  return `Matché le ${d.toLocaleDateString("fr-FR")}`;
}

async function handleUnmatch(matchId) {
  if (!confirm("Supprimer ce smash ?")) return;

  try {
    const res = await fetch(`${API_BASE}/matches/${matchId}`, {
      method: "DELETE",
      headers: { Accept: "application/json" },
    });
    if (!res.ok) return;
    fetchMatches();
  } catch (e) {
    console.error("Error unmatching:", e);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  attachEvents();
  fetchProfiles();
  fetchMatches();
});