const input = document.getElementById("search-input");
const resultsEl = document.getElementById("search-results");
const hintEl = document.getElementById("search-hint");
const clearBtn = document.getElementById("clear-btn");

let debounceTimeout = null;

function setHint(text) {
  if (!hintEl) return;
  hintEl.textContent = text;
}

function renderResults(users) {
  resultsEl.innerHTML = "";
  if (!users || users.length === 0) {
    const empty = document.createElement("div");
    empty.className = "search-list-empty";
    empty.textContent = "Aucun joueur trouvé.";
    resultsEl.appendChild(empty);
    return;
  }

  users.forEach((u) => {
    const item = document.createElement("div");
    item.className = "search-result-item";

    const avatar = document.createElement("div");
    avatar.className = "search-avatar";

    if (u.avatar && u.avatar.trim() !== "") {
      const img = document.createElement("img");
      img.src = u.avatar;
      img.alt = u.username;
      avatar.appendChild(img);
    } else {
      const span = document.createElement("span");
      span.className = "search-avatar-placeholder";
      span.textContent = u.username ? u.username[0].toUpperCase() : "?";
      avatar.appendChild(span);
    }

    const main = document.createElement("div");
    main.className = "search-user-main";

    const nameEl = document.createElement("div");
    nameEl.className = "search-username";
    nameEl.textContent = u.username;

    const metaEl = document.createElement("div");
    metaEl.className = "search-user-meta";
    metaEl.textContent = "# " + u.id;

    main.appendChild(nameEl);
    main.appendChild(metaEl);

    item.appendChild(avatar);
    item.appendChild(main);

    item.addEventListener("click", () => {
      window.location.href = "/users/" + u.id;
    });

    resultsEl.appendChild(item);
  });
}

async function searchUsers(query) {
  const trimmed = query.trim();
  if (trimmed === "") {
    resultsEl.innerHTML = "";
    setHint("Commence à taper un pseudo.");
    clearBtn.classList.remove("visible");
    return;
  }

  clearBtn.classList.add("visible");
  setHint("Recherche en cours...");

  try {
    const res = await fetch("/api/users/search?q=" + encodeURIComponent(trimmed), {
      headers: {
        "Accept": "application/json"
      },
      credentials: "include"
    });

    if (!res.ok) {
      setHint("Erreur lors de la recherche.");
      return;
    }

    const data = await res.json();
    renderResults(data);
  } catch (e) {
    setHint("Erreur réseau.");
  }
}

if (input) {
  input.addEventListener("input", () => {
    const value = input.value;
    if (debounceTimeout) {
      clearTimeout(debounceTimeout);
    }
    debounceTimeout = setTimeout(() => {
      searchUsers(value);
    }, 220);
  });
}

if (clearBtn && input) {
  clearBtn.addEventListener("click", () => {
    input.value = "";
    clearBtn.classList.remove("visible");
    resultsEl.innerHTML = "";
    setHint("Commence à taper un pseudo.");
    input.focus();
  });
}