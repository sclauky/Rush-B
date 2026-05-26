CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    birth_date DATE,
    biography TEXT,
    show_in_finder INTEGER NOT NULL DEFAULT 0 CHECK (show_in_finder IN (0, 1)),
    steam_id TEXT UNIQUE,
    two_fa_enabled INTEGER NOT NULL DEFAULT 0 CHECK (two_fa_enabled IN (0,1))
);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    session_token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS game_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    game_type_id INTEGER NOT NULL,
    FOREIGN KEY (game_type_id) REFERENCES game_types(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    game_id INTEGER NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('Chill', 'Ranked')),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_categories (
    post_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_reactions (
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    reaction_type TEXT NOT NULL CHECK (reaction_type IN ('like', 'dislike')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comment_reactions (
    comment_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    reaction_type TEXT NOT NULL CHECK (reaction_type IN ('like', 'dislike')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (comment_id, user_id),
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    games_played TEXT NOT NULL,
    play_style TEXT NOT NULL CHECK (play_style IN ('loisir', 'competition')),
    game_types TEXT NOT NULL,
    looking_for_mates INTEGER NOT NULL CHECK (looking_for_mates IN (0, 1)),
    skill_level TEXT NOT NULL,
    onboarding_done INTEGER NOT NULL DEFAULT 0 CHECK (onboarding_done IN (0, 1)),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    game_name TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_game_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    game_type TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS swipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    swiper_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    direction TEXT NOT NULL CHECK(direction IN ('like','pass')),
    created_at DATETIME NOT NULL,
    UNIQUE(swiper_id, target_id),
    FOREIGN KEY (swiper_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user1_id INTEGER NOT NULL,
    user2_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE(user1_id, user2_id),
    FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_photos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    photo_path TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS two_factor_codes (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER  NOT NULL,
    code       TEXT     NOT NULL,
    expires_at DATETIME NOT NULL,
    used       INTEGER  NOT NULL DEFAULT 0 CHECK (used IN (0,1)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT OR IGNORE INTO game_types (name) VALUES
('FPS'),
('MOBA'),
('Battle Royale'),
('MMORPG'),
('Sandbox'),
('Sport'),
('Simulation');

INSERT OR IGNORE INTO games (name, game_type_id) VALUES
('Valorant', 1),
('CS2', 1),
('League of Legends', 2),
('Fortnite', 3),
('World of Warcraft', 4),
('Minecraft', 5),
('Rocket League', 6),
('EA FC 25', 6);

INSERT OR IGNORE INTO categories (name) VALUES
('Recherche mate'),
('Discussion générale'),
('Aide'),
('Compétitif'),
('Détente'),
('Tournoi');

CREATE TABLE IF NOT EXISTS two_factor_codes (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER  NOT NULL,
    code       TEXT     NOT NULL,
    expires_at DATETIME NOT NULL,
    used       INTEGER  NOT NULL DEFAULT 0 CHECK (used IN (0,1)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS follows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    follower_id INTEGER NOT NULL,
    followed_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE(follower_id, followed_id),
    FOREIGN KEY (follower_id) REFERENCES users(id),
    FOREIGN KEY (followed_id) REFERENCES users(id)
);
SELECT u.id,
       u.username,
       COALESCE((
           SELECT photo_path
           FROM user_photos
           WHERE user_id = u.id
           ORDER BY position ASC, id ASC
           LIMIT 1
       ), '') AS avatar
FROM users u
WHERE u.id != ?
  AND u.username LIKE ?
ORDER BY u.username ASC
LIMIT 20;

-- swipe
SELECT
    u.id,
    u.username,
    u.biography,
    COALESCE((
        SELECT photo_path
        FROM user_photos
        WHERE user_id = u.id
        ORDER BY position ASC, id ASC
        LIMIT 1
    ), '') AS avatar
FROM users u
WHERE u.id != ?
  AND u.show_in_finder = 1
  AND NOT EXISTS (
      SELECT 1
      FROM swipes s
      WHERE s.swiper_id = ?
        AND s.target_id = u.id
  )
ORDER BY u.created_at DESC
LIMIT 50;
