-- Создаем базу данных, если не существует
CREATE DATABASE forum_db;

-- Подключаемся к новой базе данных
\c forum_db

-- Создаем расширение для UUID (если нужно)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица тем
CREATE TABLE IF NOT EXISTS topics (
                                      id SERIAL PRIMARY KEY,
                                      title VARCHAR(255) NOT NULL,
    description TEXT,
    user_id INTEGER REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'open',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица сообщений
CREATE TABLE IF NOT EXISTS posts (
                                     id SERIAL PRIMARY KEY,
                                     content TEXT NOT NULL,
                                     user_id INTEGER REFERENCES users(id),
    topic_id INTEGER REFERENCES topics(id),
    parent_post_id INTEGER REFERENCES posts(id) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица реакций
CREATE TABLE IF NOT EXISTS reactions (
                                         id SERIAL PRIMARY KEY,
                                         type VARCHAR(50) NOT NULL,
    user_id INTEGER REFERENCES users(id),
    post_id INTEGER REFERENCES posts(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, post_id)
    );

-- Таблица уведомлений
CREATE TABLE IF NOT EXISTS notifications (
                                             id SERIAL PRIMARY KEY,
                                             user_id INTEGER REFERENCES users(id),
    type VARCHAR(100),
    message TEXT,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица личных сообщений
CREATE TABLE IF NOT EXISTS private_messages (
                                                id SERIAL PRIMARY KEY,
                                                from_user_id INTEGER REFERENCES users(id),
    to_user_id INTEGER REFERENCES users(id),
    content TEXT,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Создаем индексы для улучшения производительности
CREATE INDEX IF NOT EXISTS idx_posts_topic_id ON posts(topic_id);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_private_messages_to_user_id ON private_messages(to_user_id);