document.addEventListener('DOMContentLoaded', () => {
  const loader = document.getElementById('page-loading-overlay');
  if (!loader) return;

  const text = loader.querySelector('.welcome-text');

  if (text) {
    const length = text.getComputedTextLength();
    text.style.strokeDasharray = String(length);
    text.style.strokeDashoffset = String(length);
    text.style.setProperty('--path-length', String(length));
    text.style.animation = 'welcome-draw 2.2s ease-in-out infinite';
  }

  const hideLoader = () => {
    window.setTimeout(() => {
      loader.classList.add('page-loading-hidden');
    }, 1200);

    window.setTimeout(() => {
      loader.style.display = 'none';
    }, 1600);
  };

  if (document.readyState === 'complete') {
    hideLoader();
  } else {
    window.addEventListener('load', hideLoader, { once: true });
  }
});