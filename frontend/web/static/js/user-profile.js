console.log("user-profile.js chargé");

function initUserProfileFollow() {
  console.log("init follow lancé");

  const btn = document.querySelector(".follow-btn");
  console.log("btn =", btn);

  if (!btn) return;

  btn.addEventListener("click", async () => {
    console.log("clic follow");

    const userId = btn.dataset.userId;
    const isFollowing = btn.dataset.following === "true";
    const url = isFollowing
      ? `/users/${userId}/unfollow`
      : `/users/${userId}/follow`;

    console.log("URL FOLLOW =", url);

    btn.disabled = true;

    try {
      const res = await fetch(url, {
        method: "POST",
        headers: {
          Accept: "application/json"
        }
      });

      console.log("STATUS FOLLOW =", res.status);

      if (!res.ok) {
        console.error("Erreur follow/unfollow");
        return;
      }

      const data = await res.json();
      const following = !!data.following;

      btn.dataset.following = following ? "true" : "false";
      btn.textContent = following ? "Followed" : "Follow";
      btn.classList.toggle("btn-primary", !following);
      btn.classList.toggle("btn-secondary", following);
    } catch (err) {
      console.error("Erreur réseau", err);
    } finally {
      btn.disabled = false;
    }
  });
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initUserProfileFollow);
} else {
  initUserProfileFollow();
}