# potbot — minimal auth demo

This project contains a minimal backend (Go) and frontend (React) implementing user registration, login, logout, and a simple post-login blank page with a navbar.

Folders:
- `backend/` — Go server that exposes REST endpoints: `/api/register`, `/api/login`, `/api/logout`, `/api/me`.
- `frontend/` — React app (CRA-style) that talks to the backend using fetch and cookies.

Database:
The project expects a MySQL database named `potbot` with a `users` table. Use the provided migration:

```sql
-- backend/MIGRATION.sql
CREATE DATABASE IF NOT EXISTS potbot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE potbot;
CREATE TABLE IF NOT EXISTS users (
  user_id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  username VARCHAR(50) NULL UNIQUE
);
```

Run instructions (macOS, zsh):

1) Prepare DB and env

Set up your MySQL server and create the `potbot` DB using the migration above. Then set environment variables for the backend (or use the defaults):

```zsh
export POTBOT_DB_DSN='root:password@tcp(127.0.0.1:3306)/potbot?parseTime=true'
export POTBOT_HASH_KEY='replace-with-32-byte-random'
export POTBOT_BLOCK_KEY='replace-with-32-byte-random'
export PORT=8080
```

2) Backend (Go)

Install Go (if you don't have it):

```zsh
brew install go
```

Then in the `backend` folder:

```zsh
cd backend
go mod tidy
go build -o potbot-server
./potbot-server
```

By default the server listens on `:8080` and serves the API and any static files if you build the frontend into `frontend/build`.

3) Frontend (React)

Install Node (if necessary):

```zsh
brew install node
```

From the `frontend` folder:

```zsh
cd frontend
npm install
npm start
```

The frontend runs on `http://localhost:3000` and expects the backend to be at `http://localhost:8080`. The frontend uses cookies (credentials included) to maintain session.

Notes and assumptions
- The backend uses a signed cookie to store a session (user_id). For production use enable Secure cookies and HTTPS.
- Passwords are stored hashed with bcrypt.
- Email uniqueness is enforced by the DB schema.

Next steps / improvements
- Add CSRF protection and HTTPS.
- Use a proper session store for server-side sessions.
- Add input validation and friendly UI / mobile styling.
