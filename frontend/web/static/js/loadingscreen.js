document.addEventListener('DOMContentLoaded', () => {
  const loader = document.getElementById('page-loading-overlay');
  if (!loader) return;

  window.addEventListener('load', () => {
    loader.classList.add('page-loading-hidden');
    window.setTimeout(() => {
      loader.style.display = 'none';
    }, 300);
  });
});