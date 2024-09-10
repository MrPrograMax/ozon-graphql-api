CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    createdBy INTEGER REFERENCES users(id) ON DELETE SET NULL,
    createdAt TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    isCommentingAvailable BOOLEAN NOT NULL
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    postId INTEGER REFERENCES posts(id) ON DELETE CASCADE,
    sender INTEGER REFERENCES users(id) ON DELETE SET NULL,
    replyTo INTEGER REFERENCES comments(id) ON DELETE SET NULL,
    text TEXT NOT NULL,
    createdAt TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (id, username) VALUES('1', 'Maxim');
INSERT INTO users (id, username) VALUES('2', 'Vika');
INSERT INTO users (id, username) VALUES('3', 'Ruslan');
