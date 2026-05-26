document.addEventListener("DOMContentLoaded", () => {
  const menuToggle = document.getElementById("menu-toggle");
  const mobileMenu = document.getElementById("mobile-menu");

  if (menuToggle && mobileMenu) {
    menuToggle.addEventListener("click", () => {
      if (mobileMenu.style.display === "flex") {
        mobileMenu.style.display = "none";
      } else {
        mobileMenu.style.display = "flex";
        mobileMenu.style.flexDirection = "column";
        mobileMenu.style.position = "absolute";
        mobileMenu.style.top = "78px";
        mobileMenu.style.left = "1rem";
        mobileMenu.style.right = "1rem";
        mobileMenu.style.padding = "1rem";
        mobileMenu.style.borderRadius = "16px";
        mobileMenu.style.background = "rgba(15, 15, 25, 0.96)";
        mobileMenu.style.border = "1px solid rgba(255,255,255,0.08)";
        mobileMenu.style.backdropFilter = "blur(16px)";
      }
    });
  }

  const registerModal = document.getElementById("registerModal");
  const openRegisterButtons = document.querySelectorAll(".open-register-modal");
  const closeRegisterButton = document.getElementById("closeRegisterModal");
  const onboardingModal = document.getElementById("onboardingModal");

  function openRegisterModal() {
    if (!registerModal) return;
    registerModal.classList.add("is-open");
    registerModal.setAttribute("aria-hidden", "false");
    document.body.classList.add("modal-open");
  }

  function closeRegisterModal() {
    if (!registerModal) return;
    registerModal.classList.remove("is-open");
    registerModal.setAttribute("aria-hidden", "true");
    document.body.classList.remove("modal-open");
  }

  openRegisterButtons.forEach((button) => {
    button.addEventListener("click", openRegisterModal);
  });

  if (closeRegisterButton) {
    closeRegisterButton.addEventListener("click", closeRegisterModal);
  }

  if (registerModal) {
    registerModal.addEventListener("click", (event) => {
      if (event.target === registerModal) {
        closeRegisterModal();
      }
    });
  }

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      closeRegisterModal();
    }
  });

  if (onboardingModal && onboardingModal.classList.contains("is-open")) {
    document.body.classList.add("modal-open");
  }

  const params = new URLSearchParams(window.location.search);
  if (params.get("modal") === "register") {
    openRegisterModal();
  }
});