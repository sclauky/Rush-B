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