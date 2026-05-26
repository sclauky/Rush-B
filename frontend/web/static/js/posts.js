/**
 * posts.js — likes/dislikes sur la page /posts (liste)
 * Place dans : frontend/web/static/js/posts.js
 */
document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll(".reaction-btn").forEach((btn) => {
    btn.addEventListener("click", async (e) => {
      // Empêche le clic de naviguer vers la page du post
      e.preventDefault();
      e.stopPropagation();

      const postID = btn.dataset.postId;
      const type = btn.dataset.type; // "like" ou "dislike"

      try {
        const res = await fetch(`/api/posts/${postID}/${type}`, {
          method: "POST",
        });
        if (!res.ok) return;
        const data = await res.json(); // { likes, dislikes, user_reaction }

        const card = btn.closest(".post-card");
        card.querySelector(".like-count").textContent = data.likes;
        card.querySelector(".dislike-count").textContent = data.dislikes;

        card
          .querySelector(".like-btn")
          .classList.toggle("active-like", data.user_reaction === "like");
        card
          .querySelector(".dislike-btn")
          .classList.toggle("active-dislike", data.user_reaction === "dislike");
      } catch (err) {
        console.error("Erreur réaction:", err);
      }
    });
  });
});
