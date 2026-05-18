(function () {
  'use strict';

  const digits     = Array.from(document.querySelectorAll('.code-digit'));
  const codeHidden = document.getElementById('code-hidden');
  const submitBtn  = document.getElementById('submit-btn');
  const errorMsg   = document.getElementById('error-msg');
  const successMsg = document.getElementById('success-msg');
  const resendBtn  = document.getElementById('resend-btn');
  const countdown  = document.getElementById('countdown');
  const progressBar  = document.getElementById('progress-bar');
  const expireLabel  = document.getElementById('expire-label');
  const tfaForm    = document.getElementById('tfa-form');

  function syncHidden() {
    const val = digits.map(d => d.value).join('');
    codeHidden.value = val;
    submitBtn.disabled = val.length < 6;
    digits.forEach(d => d.classList.toggle('filled', d.value !== ''));
  }

  function showError(msg) {
    errorMsg.textContent = msg;
    errorMsg.classList.add('visible');
    successMsg.classList.remove('visible');
    digits.forEach(d => { d.classList.add('error'); d.value = ''; });
    syncHidden();
    setTimeout(() => digits.forEach(d => d.classList.remove('error')), 600);
    digits[0].focus();
  }

  function showSuccess(msg) {
    successMsg.textContent = msg;
    successMsg.classList.add('visible');
    errorMsg.classList.remove('visible');
  }

  digits.forEach((input, i) => {
    input.addEventListener('input', () => {
      input.value = input.value.replace(/\D/g, '').slice(-1);
      syncHidden();
      if (input.value && i < digits.length - 1) digits[i + 1].focus();
    });

    input.addEventListener('keydown', e => {
      if (e.key === 'Backspace' && !input.value && i > 0) {
        digits[i - 1].value = '';
        digits[i - 1].focus();
        syncHidden();
      }
      if (e.key === 'ArrowLeft'  && i > 0)              digits[i - 1].focus();
      if (e.key === 'ArrowRight' && i < digits.length - 1) digits[i + 1].focus();
    });

    input.addEventListener('paste', e => {
      e.preventDefault();
      const pasted = (e.clipboardData || window.clipboardData)
        .getData('text')
        .replace(/\D/g, '')
        .slice(0, 6);
      pasted.split('').forEach((ch, j) => {
        if (digits[j]) digits[j].value = ch;
      });
      syncHidden();
      const nextEmpty = digits.findIndex(d => !d.value);
      (nextEmpty >= 0 ? digits[nextEmpty] : digits[digits.length - 1]).focus();
    });
  });

  digits[0].focus();

  tfaForm.addEventListener('submit', async e => {
    e.preventDefault();
    errorMsg.classList.remove('visible');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Vérification…';

    try {
      const res = await fetch('/auth/2fa/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({ code: codeHidden.value }),
        redirect: 'manual',
      });

      if (res.ok || res.status === 0 || res.type === 'opaqueredirect') {
        showSuccess('Code correct ! Redirection…');
        setTimeout(() => { location.href = '/'; }, 800);
      } else {
        const data = await res.json().catch(() => ({}));
        showError(data.error || 'Code incorrect.');
      }
    } catch {
      showError('Erreur réseau, réessaie.');
    } finally {
      submitBtn.disabled = codeHidden.value.length < 6;
      if (!successMsg.classList.contains('visible')) {
        submitBtn.textContent = 'Vérifier le code';
      }
    }
  });


  let resendTimer = null;

  function startResendCooldown(seconds) {
    resendBtn.disabled = true;
    let remaining = seconds;
    countdown.textContent = `(${remaining}s)`;
    clearInterval(resendTimer);
    resendTimer = setInterval(() => {
      remaining--;
      if (remaining <= 0) {
        clearInterval(resendTimer);
        resendBtn.disabled = false;
        countdown.textContent = '';
      } else {
        countdown.textContent = `(${remaining}s)`;
      }
    }, 1000);
  }

  startResendCooldown(60);

  resendBtn.addEventListener('click', async () => {
    resendBtn.disabled = true;
    errorMsg.classList.remove('visible');

    try {
      const res  = await fetch('/auth/2fa/resend', { method: 'POST' });
      const data = await res.json().catch(() => ({}));

      if (res.status === 429) {
        startResendCooldown(data.retry_in || 60);
      } else if (!res.ok) {
        showError(data.error || 'Impossible de renvoyer le code.');
        resendBtn.disabled = false;
      } else {
        startResendCooldown(60);
        // Confirmation visuelle
        const confirm = document.createElement('p');
        confirm.className = 'resend-confirm';
        confirm.textContent = '✓ Nouveau code envoyé !';
        resendBtn.parentNode.appendChild(confirm);
        setTimeout(() => confirm.remove(), 3000);
      }
    } catch {
      showError('Erreur réseau.');
      resendBtn.disabled = false;
    }
  });

  // ── Barre de progression (expiration 10 min = 600s) ─────────────────────────

  const TOTAL_SECONDS = 600;
  let elapsed = 0;

  const progressTimer = setInterval(() => {
    elapsed++;
    const pct = Math.max(0, 100 - (elapsed / TOTAL_SECONDS) * 100);
    progressBar.style.width = pct + '%';

    const remaining = TOTAL_SECONDS - elapsed;
    const m = Math.floor(remaining / 60);
    const s = remaining % 60;
    expireLabel.textContent = m > 0
      ? `${m} min ${String(s).padStart(2, '0')} s`
      : `${s}s`;

    if (elapsed >= TOTAL_SECONDS) {
      clearInterval(progressTimer);
      progressBar.classList.add('expired');
      expireLabel.textContent = 'Expiré';
      showError('Le code a expiré. Demande-en un nouveau.');
      submitBtn.disabled = true;
    }
  }, 1000);
})();