document.addEventListener("DOMContentLoaded", () => {
  const menuToggle = document.getElementById("menu-toggle");
  const navbar = document.querySelector(".navbar");

  if (menuToggle && navbar) {
    const closeMenu = () => {
      navbar.classList.remove("menu-open");
      menuToggle.setAttribute("aria-expanded", "false");
    };

    menuToggle.addEventListener("click", () => {
      const isOpen = navbar.classList.toggle("menu-open");
      menuToggle.setAttribute("aria-expanded", String(isOpen));
    });

    navbar.querySelectorAll(".nav-links a").forEach((link) => {
      link.addEventListener("click", closeMenu);
    });

    document.addEventListener("click", (event) => {
      if (!navbar.contains(event.target)) closeMenu();
    });

    document.addEventListener("keydown", (event) => {
      if (event.key === "Escape") closeMenu();
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

  function setupPasswordToggles() {
    document.querySelectorAll(".password-toggle").forEach((checkbox) => {
      const targets = checkbox
        .getAttribute("data-targets")
        .split(",")
        .map((selector) => selector.trim())
        .filter(Boolean);

      checkbox.addEventListener("change", () => {
        const type = checkbox.checked ? "text" : "password";
        targets.forEach((selector) => {
          const input = document.querySelector(selector);
          if (input) {
            input.type = type;
          }
        });
      });
    });
  }

  setupPasswordToggles();
});