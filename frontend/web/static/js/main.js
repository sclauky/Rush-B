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
  const closeRegisterModal = document.getElementById("closeRegisterModal");

  function openModal() {
    if (!registerModal) return;
    registerModal.classList.add("is-open");
    registerModal.setAttribute("aria-hidden", "false");
    document.body.classList.add("modal-open");
  }

  function closeModal() {
    if (!registerModal) return;
    registerModal.classList.remove("is-open");
    registerModal.setAttribute("aria-hidden", "true");
    document.body.classList.remove("modal-open");
  }

  if (openRegisterButtons.length > 0) {
    openRegisterButtons.forEach((button) => {
      button.addEventListener("click", openModal);
    });
  }

  if (closeRegisterModal) {
    closeRegisterModal.addEventListener("click", closeModal);
  }

  if (registerModal) {
    registerModal.addEventListener("click", (event) => {
      if (event.target === registerModal) {
        closeModal();
      }
    });
  }

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      closeModal();
    }
  });
});

document.addEventListener("DOMContentLoaded", () => {
  const onboardingModal = document.getElementById("onboardingModal");

  if (onboardingModal && onboardingModal.classList.contains("is-open")) {
    document.body.classList.add("modal-open");
  }
});