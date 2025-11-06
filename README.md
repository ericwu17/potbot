# potbot

This project contains a minimal backend (Go) and frontend (React) implementing user registration, login, logout, and a simple post-login blank page with a navbar.

Folders:
- `backend/`. Go server that exposes REST endpoints: `/api/register`, `/api/login`, `/api/logout`, `/api/me`.
- `frontend/`. React app (CRA-style) that talks to the backend using fetch and cookies.

## Running the app

In the `backend` folder:

```bash
cd backend
go mod tidy
go run .
```

The backend folder also requires a `.env` file that has contents in the format of `.env.template`.

By default the server listens on `:8080`.

In the `frontend` folder:

```bash
cd frontend
npm install
npm start
```

The frontend runs on `http://localhost:3000` and expects the backend to be at `http://localhost:8080`. The frontend uses cookies to maintain session.

## Database

The backend will try to connect to a mysql database with a database named `potbot`.

The current schema is below, which is subject to change:

```sql
mysql> desc plants
    -> ;
+------------+--------------+------+-----+---------+----------------+
| Field      | Type         | Null | Key | Default | Extra          |
+------------+--------------+------+-----+---------+----------------+
| id         | int          | NO   | PRI | NULL    | auto_increment |
| user_id    | int          | NO   |     | NULL    |                |
| plant_id   | varchar(100) | NO   | UNI | NULL    |                |
| plant_type | varchar(100) | NO   |     | NULL    |                |
+------------+--------------+------+-----+---------+----------------+
4 rows in set (0.01 sec)

mysql> desc users;
+---------------+--------------+------+-----+---------+----------------+
| Field         | Type         | Null | Key | Default | Extra          |
+---------------+--------------+------+-----+---------+----------------+
| user_id       | int          | NO   | PRI | NULL    | auto_increment |
| email         | varchar(255) | NO   | UNI | NULL    |                |
| password_hash | varchar(255) | NO   |     | NULL    |                |
| username      | varchar(50)  | YES  | UNI | NULL    |                |
+---------------+--------------+------+-----+---------+----------------+
4 rows in set (0.00 sec)

mysql> show tables;
+------------------+
| Tables_in_potbot |
+------------------+
| plants           |
| users            |
+------------------+
2 rows in set (0.00 sec)

```

