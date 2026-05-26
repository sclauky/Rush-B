function initNavbarSearch() {
  const toggle = document.querySelector(".nav-search-toggle");
  const panel = document.querySelector(".nav-search-panel");
  const input = document.getElementById("navSearchInput");
  const resultsContainer = document.getElementById("navSearchResults");

  console.log("init navbar search", { toggle, panel, input, resultsContainer });

  if (!toggle || !panel || !input || !resultsContainer) {
    return;
  }

  let lastController = null;

  function closePanel() {
    panel.setAttribute("data-open", "false");
    panel.classList.remove("nav-search-panel--open");
    resultsContainer.innerHTML = "";
    input.value = "";
  }

  toggle.addEventListener("click", (e) => {
    e.preventDefault();
    e.stopPropagation();
    console.log("click loupe");

    panel.classList.toggle("nav-search-panel--open");

    const isOpen = panel.classList.contains("nav-search-panel--open");
    panel.setAttribute("data-open", isOpen ? "true" : "false");

    if (isOpen) {
      setTimeout(() => input.focus(), 50);
    }
  });

  panel.addEventListener("click", (e) => {
    e.stopPropagation();
  });

  document.addEventListener("click", (e) => {
    if (!panel.contains(e.target) && !toggle.contains(e.target)) {
      closePanel();
    }
  });

  input.addEventListener("keydown", (e) => {
    if (e.key === "Escape") {
      e.preventDefault();
      closePanel();
    }
  });

  input.addEventListener("input", async () => {
    const q = input.value.trim();

    if (q.length < 2) {
      resultsContainer.innerHTML = "";
      return;
    }

    if (lastController) {
      lastController.abort();
    }
    lastController = new AbortController();

    resultsContainer.innerHTML =
      '<div class="nav-search-item nav-search-item--loading">Recherche…</div>';

    try {
      const res = await fetch(`/api/users/search?q=${encodeURIComponent(q)}`, {
        headers: { Accept: "application/json" },
        signal: lastController.signal,
      });

      if (!res.ok) {
        resultsContainer.innerHTML =
          '<div class="nav-search-item nav-search-item--empty">Erreur de recherche</div>';
        return;
      }

      const data = await res.json();
      const users = data.users || [];

      if (users.length === 0) {
        resultsContainer.innerHTML =
          '<div class="nav-search-item nav-search-item--empty">Aucun joueur trouvé</div>';
        return;
      }

      resultsContainer.innerHTML = users.map((u) => `
        <a class="nav-search-item" href="/users/${u.id}">
          <div class="nav-search-avatar">
            ${
              u.avatar_path
                ? `<img src="${u.avatar_path}" alt="${u.username}">`
                : `<div class="nav-search-avatar-fallback">${u.username?.[0]?.toUpperCase() || "?"}</div>`
            }
          </div>
          <div class="nav-search-meta">
            <div class="nav-search-username">${u.username}</div>
            ${u.biography ? `<div class="nav-search-bio">${u.biography}</div>` : ""}
          </div>
        </a>
      `).join("");
    } catch (err) {
      if (err.name === "AbortError") return;
      resultsContainer.innerHTML =
        '<div class="nav-search-item nav-search-item--empty">Erreur réseau</div>';
    }
  });
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initNavbarSearch);
} else {
  initNavbarSearch();
}