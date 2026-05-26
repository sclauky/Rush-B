/**
 * post-detail.js
 * Gère les likes/dislikes et commentaires sur la page /posts/:id
 * Place dans : frontend/web/static/js/post-detail.js
 */

document.addEventListener("DOMContentLoaded", () => {
  const POST_ID = window.POST_ID;
  if (!POST_ID) return;

  // ──────────────────────────────────────────────
  // LIKES / DISLIKES sur le post
  // ──────────────────────────────────────────────
  document.querySelectorAll(".reaction-btn").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const type = btn.dataset.type;
      try {
        const res = await fetch(`/api/posts/${POST_ID}/${type}`, {
          method: "POST",
        });
        if (!res.ok) return;
        const data = await res.json(); // { likes, dislikes, user_reaction }

        document.querySelector(".like-count").textContent = data.likes;
        document.querySelector(".dislike-count").textContent = data.dislikes;

        document
          .querySelector(".like-btn")
          .classList.toggle("active-like", data.user_reaction === "like");
        document
          .querySelector(".dislike-btn")
          .classList.toggle("active-dislike", data.user_reaction === "dislike");
      } catch (err) {
        console.error("Réaction post:", err);
      }
    });
  });

  // ──────────────────────────────────────────────
  // CHARGEMENT DES COMMENTAIRES
  // ──────────────────────────────────────────────
  loadComments();

  async function loadComments() {
    const list = document.getElementById("comment-list");
    try {
      const res = await fetch(`/api/posts/${POST_ID}/comments`);
      if (!res.ok) throw new Error("fetch failed");
      const comments = await res.json();
      renderComments(comments || []);
    } catch (err) {
      list.innerHTML = `<p class="comments-empty">Impossible de charger les commentaires.</p>`;
      console.error("loadComments:", err);
    }
  }

  function renderComments(comments) {
    const list = document.getElementById("comment-list");
    if (!comments.length) {
      list.innerHTML = `<p class="comments-empty">Aucun commentaire pour l'instant. Sois le premier !</p>`;
      return;
    }
    list.innerHTML = comments.map(buildCommentHTML).join("");
    attachCommentEvents();
  }

  function buildCommentHTML(c) {
    const likeClass = c.user_reaction === "like" ? "active-like" : "";
    const dislikeClass = c.user_reaction === "dislike" ? "active-dislike" : "";

    const date = new Date(c.created_at);
    const dateStr = isNaN(date.getTime())
      ? ""
      : date.toLocaleDateString("fr-FR", {
          day: "2-digit",
          month: "2-digit",
          year: "numeric",
          hour: "2-digit",
          minute: "2-digit",
        });

    return `
      <div class="comment-item" data-comment-id="${c.id}">
        <div class="comment-header">
          <span class="comment-author">${escapeHTML(c.author_name)}</span>
          <span class="comment-date">${dateStr}</span>
          <div class="comment-own-actions">
            <button class="comment-action-btn edit-btn"   data-comment-id="${c.id}" title="Modifier">✏️</button>
            <button class="comment-action-btn delete delete-btn" data-comment-id="${c.id}" title="Supprimer">🗑️</button>
          </div>
        </div>

        <div class="comment-content" id="cc-${c.id}">${escapeHTML(c.content)}</div>

        <div class="comment-reactions">
          <button class="comment-reaction-btn ${likeClass}"    data-comment-id="${c.id}" data-type="like">
            👍 <span class="cl-${c.id}">${c.like_count}</span>
          </button>
          <button class="comment-reaction-btn ${dislikeClass}" data-comment-id="${c.id}" data-type="dislike">
            👎 <span class="cd-${c.id}">${c.dislike_count}</span>
          </button>
        </div>
      </div>
    `;
  }

  function attachCommentEvents() {
    const list = document.getElementById("comment-list");

    list.querySelectorAll(".comment-reaction-btn").forEach((btn) => {
      btn.addEventListener("click", () => reactToComment(btn));
    });
    list.querySelectorAll(".delete-btn").forEach((btn) => {
      btn.addEventListener("click", () => deleteComment(btn));
    });
    list.querySelectorAll(".edit-btn").forEach((btn) => {
      btn.addEventListener("click", () => showEditForm(btn));
    });
  }

  // ──────────────────────────────────────────────
  // POSTER un commentaire
  // ──────────────────────────────────────────────
  const submitBtn = document.getElementById("new-comment-submit");
  const textarea = document.getElementById("new-comment-input");

  submitBtn.addEventListener("click", submitComment);
  textarea.addEventListener("keydown", (e) => {
    if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) submitComment();
  });
  textarea.addEventListener("input", () => {
    textarea.style.height = "auto";
    textarea.style.height = Math.min(textarea.scrollHeight, 160) + "px";
  });

  async function submitComment() {
    const content = textarea.value.trim();
    if (!content) return;

    submitBtn.disabled = true;
    try {
      const body = new URLSearchParams({ content });
      const res = await fetch(`/api/posts/${POST_ID}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body,
      });
      if (!res.ok) throw new Error("submit failed");

      textarea.value = "";
      textarea.style.height = "auto";
      await loadComments();

      // Met à jour le compteur en haut
      const total = document.getElementById("comment-total");
      if (total) total.textContent = parseInt(total.textContent || "0") + 1;
    } catch (err) {
      console.error("submitComment:", err);
    } finally {
      submitBtn.disabled = false;
    }
  }

  // ──────────────────────────────────────────────
  // RÉACTIONS sur commentaire
  // ──────────────────────────────────────────────
  async function reactToComment(btn) {
    const commentID = btn.dataset.commentId;
    const type = btn.dataset.type;

    try {
      const res = await fetch(
        `/api/posts/${POST_ID}/comments/${commentID}/${type}`,
        { method: "POST" },
      );
      if (!res.ok) return;
      const data = await res.json();

      document.querySelector(`.cl-${commentID}`).textContent = data.likes;
      document.querySelector(`.cd-${commentID}`).textContent = data.dislikes;

      const item = btn.closest(".comment-item");
      item.querySelectorAll(".comment-reaction-btn").forEach((b) => {
        b.classList.remove("active-like", "active-dislike");
      });
      if (data.user_reaction === "like") {
        item.querySelector(`[data-type="like"]`).classList.add("active-like");
      } else if (data.user_reaction === "dislike") {
        item
          .querySelector(`[data-type="dislike"]`)
          .classList.add("active-dislike");
      }
    } catch (err) {
      console.error("reactToComment:", err);
    }
  }

  // ──────────────────────────────────────────────
  // SUPPRIMER un commentaire
  // ──────────────────────────────────────────────
  async function deleteComment(btn) {
    const commentID = btn.dataset.commentId;
    if (!confirm("Supprimer ce commentaire ?")) return;

    try {
      const res = await fetch(`/api/posts/${POST_ID}/comments/${commentID}`, {
        method: "DELETE",
      });
      if (res.ok) {
        btn.closest(".comment-item").remove();
        const total = document.getElementById("comment-total");
        if (total)
          total.textContent = Math.max(
            0,
            parseInt(total.textContent || "0") - 1,
          );

        const list = document.getElementById("comment-list");
        if (!list.querySelector(".comment-item")) {
          list.innerHTML = `<p class="comments-empty">Aucun commentaire pour l'instant. Sois le premier !</p>`;
        }
      } else if (res.status === 403) {
        alert("Tu ne peux supprimer que tes propres commentaires.");
      }
    } catch (err) {
      console.error("deleteComment:", err);
    }
  }

  // ──────────────────────────────────────────────
  // ÉDITER un commentaire (inline)
  // ──────────────────────────────────────────────
  function showEditForm(btn) {
    const commentID = btn.dataset.commentId;
    const contentDiv = document.getElementById(`cc-${commentID}`);
    const current = contentDiv.textContent.trim();

    if (document.getElementById(`ef-${commentID}`)) return; // déjà ouvert

    const form = document.createElement("div");
    form.className = "comment-edit-form";
    form.id = `ef-${commentID}`;
    form.innerHTML = `
      <input class="comment-edit-input" type="text" value="${escapeHTML(current)}" />
      <button class="comment-edit-save">Sauver</button>
      <button class="comment-edit-cancel">Annuler</button>
    `;

    contentDiv.after(form);
    form.querySelector(".comment-edit-input").focus();

    form
      .querySelector(".comment-edit-cancel")
      .addEventListener("click", () => form.remove());
    form
      .querySelector(".comment-edit-save")
      .addEventListener("click", async () => {
        const newContent = form
          .querySelector(".comment-edit-input")
          .value.trim();
        if (!newContent) return;

        try {
          const body = new URLSearchParams({ content: newContent });
          const res = await fetch(
            `/api/posts/${POST_ID}/comments/${commentID}`,
            {
              method: "PUT",
              headers: { "Content-Type": "application/x-www-form-urlencoded" },
              body,
            },
          );
          if (res.ok) {
            contentDiv.textContent = newContent;
            form.remove();
          } else if (res.status === 403) {
            alert("Tu ne peux modifier que tes propres commentaires.");
          }
        } catch (err) {
          console.error("updateComment:", err);
        }
      });
  }

  // ──────────────────────────────────────────────
  // UTILITAIRE
  // ──────────────────────────────────────────────
  function escapeHTML(str) {
    return String(str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }
});
