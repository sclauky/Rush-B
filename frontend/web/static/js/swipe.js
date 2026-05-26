 const API = '';
    let profiles = [];
    let currentIdx = 0;
    let currentCard = null;
    let isDragging = false;
    let startX = 0, startY = 0, currentX = 0;
    let matches = [];
    let currentPhotoIdx = 0;
    let myInitial = 'M';

    // Game type → chip class
    function chipClass(type) {
      const t = type.toLowerCase();
      if (t.includes('fps') || t.includes('tir') || t.includes('shoot')) return 'chip-fps';
      if (t.includes('rpg') || t.includes('rôle') || t.includes('role')) return 'chip-rpg';
      if (t.includes('moba')) return 'chip-moba';
      if (t.includes('battle') || t.includes('royale')) return 'chip-battle';
      if (t.includes('sport')) return 'chip-sport';
      if (t.includes('strat') || t.includes('rts')) return 'chip-strat';
      return 'chip-other';
    }

    function gameTypeIcon(type) {
      const t = type.toLowerCase();
      if (t.includes('fps') || t.includes('tir') || t.includes('shoot')) return '🎯';
      if (t.includes('rpg') || t.includes('rôle') || t.includes('role')) return '⚔️';
      if (t.includes('moba')) return '🗡️';
      if (t.includes('battle') || t.includes('royale')) return '🪂';
      if (t.includes('sport')) return '⚽';
      if (t.includes('strat') || t.includes('rts')) return '🗺️';
      if (t.includes('plateforme') || t.includes('platform')) return '🏃';
      if (t.includes('horror') || t.includes('horreur')) return '👻';
      if (t.includes('sim')) return '🌍';
      return '🎮';
    }

    function skillBadge(level) {
      const map = { noob: '🥉', casual: '🥈', regular: '🥇', hardcore: '💎', pro: '👑' };
      const lc = level.toLowerCase();
      for (const [k, v] of Object.entries(map)) {
        if (lc.includes(k)) return v + ' ' + level;
      }
      return '🎮 ' + level;
    }

    function buildCard(user) {
      const card = document.createElement('div');
      card.className = 'swipe-card';
      card.dataset.userId = user.id;

      const hasPhotos = user.photos && user.photos.length > 0;
      const photos = user.photos || [];

      // Photo HTML
      let photosHTML = '';
      if (hasPhotos) {
        photosHTML = photos.map((p, i) =>
          `<div class="card-photo ${i === 0 ? 'active' : ''}" style="background-image:url('${p}')"></div>`
        ).join('');
      } else {
        photosHTML = `<div class="card-photo-default active">
          <span class="avatar-text">${user.first_name.charAt(0)}</span>
        </div>`;
      }

      const dotsHTML = hasPhotos && photos.length > 1 ?
        `<div class="photo-dots">${photos.map((_, i) =>
          `<div class="photo-dot ${i === 0 ? 'active' : ''}"></div>`).join('')}</div>` : '';

      const navHTML = hasPhotos && photos.length > 1 ? `
        <div class="photo-nav photo-nav-prev" data-dir="-1">
          <div class="photo-nav-btn">‹</div>
        </div>
        <div class="photo-nav photo-nav-next" data-dir="1">
          <div class="photo-nav-btn">›</div>
        </div>` : '';

      const gameTypesHTML = (user.game_types || []).map(gt =>
        `<span class="game-type-chip ${chipClass(gt)}">${gameTypeIcon(gt)} ${gt}</span>`
      ).join('');

      const gamesHTML = (user.games_played || []).slice(0, 6).map(g =>
        `<span class="card-game-pill">${g}</span>`
      ).join('');

      const metaHTML = [
        user.skill_level ? `<span class="card-meta-tag">${skillBadge(user.skill_level)}</span>` : '',
        user.play_style ? `<span class="card-meta-tag">${user.play_style === 'competition' ? '🏆 Compétitif' : '😎 Casual'}</span>` : ''
      ].filter(Boolean).join('');

      card.innerHTML = `
        <div class="stamp stamp-like">LIKE</div>
        <div class="stamp stamp-pass">PASS</div>
        <div class="card-photos" id="card-photos">
          ${photosHTML}
          ${navHTML}
          ${dotsHTML}
          <div class="photo-gradient"></div>
        </div>
        <div class="card-body">
          <div class="card-name">
            ${user.first_name} ${user.last_name}
            <span class="card-username">@${user.username}</span>
          </div>
          ${gameTypesHTML ? `<div class="card-game-types">${gameTypesHTML}</div>` : ''}
          ${metaHTML ? `<div class="card-meta-row">${metaHTML}</div>` : ''}
          ${user.biography ? `
            <div class="card-divider"></div>
            <div class="card-section-title">À propos</div>
            <p class="card-bio">${user.biography}</p>
          ` : ''}
          ${gamesHTML ? `
            <div class="card-divider"></div>
            <div class="card-section-title">Jeux joués</div>
            <div class="card-games">${gamesHTML}</div>
          ` : ''}
        </div>
      `;

      // Photo navigation
      let photoIdx = 0;
      function setPhoto(idx) {
        const photoDivs = card.querySelectorAll('.card-photo');
        const dots = card.querySelectorAll('.photo-dot');
        if (!photoDivs.length) return;
        photoIdx = (idx + photoDivs.length) % photoDivs.length;
        photoDivs.forEach((p, i) => p.classList.toggle('active', i === photoIdx));
        dots.forEach((d, i) => d.classList.toggle('active', i === photoIdx));
      }

      card.querySelectorAll('.photo-nav').forEach(nav => {
        nav.addEventListener('click', e => {
          e.stopPropagation();
          setPhoto(photoIdx + parseInt(nav.dataset.dir));
        });
      });

      // Drag to swipe
      setupDrag(card);
      return card;
    }

    function setupDrag(card) {
      let startX, startY, dragging = false;

      const onStart = e => {
        if (e.target.closest('.photo-nav') || e.target.closest('.card-body')) {
          dragging = false; return;
        }
        dragging = true;
        startX = (e.touches ? e.touches[0].clientX : e.clientX);
        startY = (e.touches ? e.touches[0].clientY : e.clientY);
        card.classList.add('is-dragging');
      };
      const onMove = e => {
        if (!dragging) return;
        const x = (e.touches ? e.touches[0].clientX : e.clientX) - startX;
        const y = (e.touches ? e.touches[0].clientY : e.clientY) - startY;
        const rot = x * 0.08;
        card.style.transform = `translateX(${x}px) translateY(${y * 0.3}px) rotate(${rot}deg)`;
        if (x > 50) card.classList.add('hint-like');
        else if (x < -50) card.classList.add('hint-pass');
        else { card.classList.remove('hint-like', 'hint-pass'); }
      };
      const onEnd = e => {
        if (!dragging) return;
        dragging = false;
        card.classList.remove('is-dragging', 'hint-like', 'hint-pass');
        const x = parseInt(card.style.transform.match(/translateX\((-?\d+)/) ?.[1] || 0);
        if (x > 100) doSwipe('like');
        else if (x < -100) doSwipe('pass');
        else {
          card.style.transform = '';
        }
      };

      card.addEventListener('mousedown', onStart);
      document.addEventListener('mousemove', onMove);
      document.addEventListener('mouseup', onEnd);
      card.addEventListener('touchstart', onStart, { passive: true });
      card.addEventListener('touchmove', onMove, { passive: true });
      card.addEventListener('touchend', onEnd);
    }

    function showCard() {
      if (currentIdx >= profiles.length) {
        document.getElementById('empty-state').classList.add('visible');
        document.getElementById('swipe-actions').style.display = 'none';
        return;
      }
      document.getElementById('empty-state').classList.remove('visible');
      document.getElementById('swipe-actions').style.display = 'flex';

      if (currentCard) currentCard.remove();
      currentCard = buildCard(profiles[currentIdx]);
      document.getElementById('card-stack').appendChild(currentCard);
    }

    async function doSwipe(direction) {
      if (currentIdx >= profiles.length) return;
      const user = profiles[currentIdx];
      currentIdx++;

      if (direction === 'like') currentCard?.classList.add('fly-right');
      else currentCard?.classList.add('fly-left');

      setTimeout(() => {
        currentCard?.remove();
        currentCard = null;
        showCard();
      }, 450);

      try {
        const res = await fetch(`${API}/api/swipe/action`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ target_id: user.id, direction })
        });
        const data = await res.json();
        if (data.matched) showMatchModal(user, data.match_id);
      } catch (e) { console.error(e); }
    }

    async function loadProfiles() {
      document.getElementById('loading-state').style.display = 'flex';
      document.getElementById('empty-state').classList.remove('visible');
      document.getElementById('swipe-actions').style.display = 'none';
      if (currentCard) { currentCard.remove(); currentCard = null; }

      try {
        const res = await fetch(`${API}/api/swipe/profiles`, { credentials: 'include' });
        if (res.status === 401) { location.href = '/login.html'; return; }
        profiles = await res.json();
        currentIdx = 0;
      } catch(e) { profiles = []; }

      document.getElementById('loading-state').style.display = 'none';
      showCard();
    }

    async function loadMatches() {
      try {
        const res = await fetch(`${API}/api/swipe/matches`, { credentials: 'include' });
        matches = await res.json();
      } catch(e) { matches = []; }
      renderMatchesSidebar();
      renderMatchesFull();
      const cnt = matches.length;
      document.getElementById('match-sidebar-count').textContent = cnt;
      document.getElementById('match-tab-count').textContent = cnt > 0 ? ` (${cnt})` : '';
    }

    function renderMatchesSidebar() {
      const el = document.getElementById('matches-sidebar-list');
      if (!matches.length) {
        el.innerHTML = `<div class="matches-empty"><div class="matches-empty-icon">🎯</div><p>Tes matchs apparaîtront ici.</p></div>`;
        return;
      }
      el.innerHTML = matches.slice(0, 8).map(m => {
        const u = m.user;
        const photo = u.photos?.[0];
        const avatarHTML = photo
          ? `<img src="${photo}" alt="">`
          : `<span class="match-avatar-text">${u.first_name.charAt(0)}</span>`;
        const tagsHTML = (u.game_types || []).slice(0, 2).map(t =>
          `<span class="match-tag">${t}</span>`).join('');
        return `<div class="match-card">
          <div class="match-avatar">${avatarHTML}</div>
          <div class="match-info">
            <div class="match-name">${u.first_name} ${u.last_name}</div>
            <div class="match-tags">${tagsHTML}</div>
          </div>
          <div class="match-new"></div>
        </div>`;
      }).join('');
    }

    function renderMatchesFull() {
      const el = document.getElementById('matches-full-grid');
      if (!matches.length) {
        el.innerHTML = `<div class="matches-empty" style="grid-column:1/-1;padding:4rem 0;">
          <div class="matches-empty-icon">🎮</div><p>Aucun match pour le moment.</p></div>`;
        return;
      }
      el.innerHTML = matches.map(m => {
        const u = m.user;
        const photo = u.photos?.[0];
        const avatarHTML = photo
          ? `<img src="${photo}" alt="" style="width:100%;height:100%;object-fit:cover;">`
          : `<span style="font-family:'Orbitron';font-size:1.4rem;font-weight:900;background:linear-gradient(135deg,var(--primary),var(--primary-2));-webkit-background-clip:text;-webkit-text-fill-color:transparent;">${u.first_name.charAt(0)}</span>`;
        const tagsHTML = (u.game_types || []).map(t =>
          `<span class="game-type-chip ${chipClass(t)}" style="font-size:.72rem;">${gameTypeIcon(t)} ${t}</span>`).join('');
        const date = new Date(m.matched_at).toLocaleDateString('fr-FR', { day:'numeric', month:'long' });
        return `<div class="match-full-card">
          <div class="match-full-top">
            <div class="match-full-avatar">${avatarHTML}</div>
            <div>
              <div class="match-full-name">${u.first_name} ${u.last_name}</div>
              <div class="match-full-username">@${u.username}</div>
              <div class="match-full-date">Matché le ${date}</div>
            </div>
          </div>
          <div class="match-full-tags">${tagsHTML}</div>
          <div class="match-full-actions">
            <button class="btn btn-ghost btn-sm">💬 Message</button>
            <button class="btn btn-unmatch btn-sm" onclick="unmatch(${m.match_id})">Unmatch</button>
          </div>
        </div>`;
      }).join('');
    }

    async function unmatch(matchId) {
      if (!confirm('Défaire ce match ?')) return;
      await fetch(`${API}/api/swipe/matches/${matchId}`, { method: 'DELETE', credentials: 'include' });
      await loadMatches();
    }

    function showMatchModal(user, matchId) {
      document.getElementById('match-them-text').textContent = user.first_name.charAt(0);
      document.getElementById('match-me-text').textContent = myInitial;
      const photo = user.photos?.[0];
      if (photo) {
        document.getElementById('match-them-avatar').innerHTML = `<img src="${photo}" style="width:100%;height:100%;object-fit:cover;border-radius:18px;">`;
      }
      document.getElementById('match-modal').classList.add('visible');
    }

    function closeMatchModal() {
      document.getElementById('match-modal').classList.remove('visible');
      loadMatches();
    }

    function viewMatches() {
      closeMatchModal();
      document.getElementById('tab-matches').click();
    }
    
    document.getElementById('tab-swipe').addEventListener('click', () => {
      document.getElementById('tab-swipe').classList.add('active');
      document.getElementById('tab-matches').classList.remove('active');
      document.getElementById('view-swipe').style.display = '';
      document.getElementById('view-matches').classList.remove('visible');
    });
    document.getElementById('tab-matches').addEventListener('click', () => {
      document.getElementById('tab-matches').classList.add('active');
      document.getElementById('tab-swipe').classList.remove('active');
      document.getElementById('view-swipe').style.display = 'none';
      document.getElementById('view-matches').classList.add('visible');
      loadMatches();
    });


    document.getElementById('btn-like').addEventListener('click', () => doSwipe('like'));
    document.getElementById('btn-pass').addEventListener('click', () => doSwipe('pass'));
    document.getElementById('btn-super').addEventListener('click', () => doSwipe('like'));

  
    document.addEventListener('keydown', e => {
      if (e.key === 'ArrowRight') doSwipe('like');
      if (e.key === 'ArrowLeft') doSwipe('pass');
    });

    loadProfiles();
    loadMatches();