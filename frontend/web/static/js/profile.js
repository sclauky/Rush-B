// Charger les informations du profil
async function loadProfile() {
    try {
        const userId = new URLSearchParams(window.location.search).get('id') || '1';
        const response = await fetch(`/api/user/${userId}`);

        if (!response.ok) {
            throw new Error('Erreur lors du chargement du profil');
        }

        const user = await response.json();

        // Mettre à jour les éléments de la page
        document.getElementById('profileAvatar').src = user.avatar || '/uploads/default-avatar.jpg';
        document.getElementById('profileUsername').textContent = user.username;
        document.getElementById('profileBio').textContent = user.bio || 'Aucune bio disponible';
        document.getElementById('profileEmail').textContent = user.email;
        document.getElementById('postCount').textContent = user.post_count;
        document.getElementById('followersCount').textContent = user.followers;

        // Formater la date
        const joinDate = new Date(user.created_at);
        const options = { year: 'numeric', month: 'long' };
        document.getElementById('profileJoined').textContent = 
            `Membre depuis ${joinDate.toLocaleDateString('fr-FR', options)}`;

        // Charger les posts de l'utilisateur
        loadUserPosts(userId);
    } catch (error) {
        console.error('Erreur:', error);
        document.getElementById('profileUsername').textContent = 'Erreur de chargement';
    }
}

// Charger les posts de l'utilisateur
async function loadUserPosts(userId) {
    try {
        const response = await fetch(`/api/user/posts/${userId}`);

        if (!response.ok) {
            throw new Error('Erreur lors du chargement des posts');
        }

        const posts = await response.json();
        const postsContainer = document.getElementById('userPosts');

        if (posts.length === 0) {
            postsContainer.innerHTML = '<div class="loading-message">Aucun post pour le moment</div>';
            return;
        }

        postsContainer.innerHTML = posts.map(post => {
            const postDate = new Date(post.created_at);
            const relativeTime = getRelativeTime(postDate);

            return `
                <div class="post-card">
                    <p class="post-content">${escapeHtml(post.content)}</p>
                    ${post.image_url ? `<img src="${post.image_url}" alt="Post image" style="width: 100%; border-radius: 12px; margin: 1rem 0;">` : ''}
                    <div class="post-footer">
                        <span class="post-date">${relativeTime}</span>
                        <div class="post-likes">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
                            </svg>
                            ${post.likes}
                        </div>
                    </div>
                </div>
            `;
        }).join('');

    } catch (error) {
        console.error('Erreur:', error);
        document.getElementById('userPosts').innerHTML = 
            '<div class="loading-message">Erreur lors du chargement des posts</div>';
    }
}

// Fonction utilitaire pour obtenir le temps relatif
function getRelativeTime(date) {
    const now = new Date();
    const diffInSeconds = Math.floor((now - date) / 1000);

    if (diffInSeconds < 60) return "À l'instant";
    if (diffInSeconds < 3600) return `Il y a ${Math.floor(diffInSeconds / 60)} min`;
    if (diffInSeconds < 86400) return `Il y a ${Math.floor(diffInSeconds / 3600)}h`;
    if (diffInSeconds < 604800) return `Il y a ${Math.floor(diffInSeconds / 86400)}j`;

    return date.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short', year: 'numeric' });
}

// Fonction utilitaire pour échapper le HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Charger le profil au chargement de la page
document.addEventListener('DOMContentLoaded', loadProfile);
