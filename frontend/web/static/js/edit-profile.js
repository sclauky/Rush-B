document.addEventListener("DOMContentLoaded", () => {
  const bioInput = document.getElementById("bio");
  const bioCount = document.getElementById("bioCount");
  const avatarInput = document.getElementById("avatarInput");
  const avatarPreview = document.getElementById("avatarPreview");

  function updateBioCount() {
    if (!bioInput || !bioCount) return;

    if (bioInput.value.length > 200) {
      bioInput.value = bioInput.value.slice(0, 200);
    }

    bioCount.textContent = bioInput.value.length;
  }

  if (bioInput) {
    updateBioCount();
    bioInput.addEventListener("input", updateBioCount);
    bioInput.addEventListener("paste", () => {
      setTimeout(updateBioCount, 0);
    });
  }

  if (avatarInput && avatarPreview) {
    avatarInput.addEventListener("change", (event) => {
      const file = event.target.files && event.target.files[0];
      if (!file) return;

      const reader = new FileReader();
      reader.onload = (e) => {
        avatarPreview.src = e.target.result;
      };
      reader.readAsDataURL(file);
    });
  }
});