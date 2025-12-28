# 📘 Group Bill Splitter - Schema Documentation

This document outlines the **High-Level Design (HLD)** and **Low-Level Design (LLD)** of the `Group Bill Splitter` system, describing models, features, database schema, and entity relationships.

---

## ✅ Features

- 👤 User Registration, Activation, Login
- 👥 Group Creation, Update, Deletion
- 🧾 Add/Update/Delete Bills in Groups
- 📊 Bill Splitting Calculation (Per Head)
- 🔁 Recalculation of Splits
- 🔐 Group-level Permissions (View, Edit, Create, Delete)
- 📨 OTP-based Verification (Activation / Reset Password)

---

## 🧠 High-Level Design (HLD)

```text
                   +----------------+         +----------------+
                   |    Frontend    | <-----> |    Backend     |
                   +----------------+         +----------------+
                                                    |
                                                    v
                                     +-----------------------------+
                                     |         API Layer           |
                                     +-----------------------------+
                                                    |
                                                    v
                      +----------------+     +---------------------+
                      |   Services     | <-- |  Middleware (Auth)  |
                      +----------------+     +---------------------+
                              |
                              v
                     +-------------------+
                     |     Repository    |
                     +-------------------+
                              |
                              v
                    +----------------------+
                    |   PostgreSQL DB      |
                    +----------------------+
```

---

## ⚙️ Low-Level Design (LLD)

### 🧍‍♂️ User
```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100),
  email VARCHAR(100) UNIQUE NOT NULL,
  password TEXT NOT NULL,
  is_active BOOLEAN DEFAULT FALSE,
  created_by TEXT,
  updated_by TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
CREATE UNIQUE INDEX idx_users_email ON users(email);
```

### 🔐 AuthToken
```sql
CREATE TABLE auth_tokens (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  access_token TEXT NOT NULL,
  refresh_token TEXT NOT NULL,
  access_expires_at TIMESTAMP,
  refresh_expires_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
```

### 📩 OTP
```sql
CREATE TABLE otps (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  code VARCHAR(10),
  purpose VARCHAR(50), -- activation | password_reset
  expires_at TIMESTAMP,
  used BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
```

### 🧑‍🤝‍🧑 Group
```sql
CREATE TABLE groups (
  id SERIAL PRIMARY KEY,
  owner_id INT REFERENCES users(id),
  name VARCHAR(100),
  description TEXT,
  created_by TEXT,
  updated_by TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
```

### ✅ GroupUserPermissions
```sql
CREATE TABLE group_user_permissions (
  id SERIAL PRIMARY KEY,
  group_id INT REFERENCES groups(id),
  user_id INT REFERENCES users(id),
  permission_type VARCHAR(20), -- view | edit | create | delete
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
CREATE INDEX idx_permissions_group_user ON group_user_permissions(group_id, user_id);
```

### 💸 Bill
```sql
CREATE TABLE bills (
  id SERIAL PRIMARY KEY,
  group_id INT REFERENCES groups(id),
  user_id INT REFERENCES users(id),
  paid_amount FLOAT NOT NULL,
  description TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
CREATE INDEX idx_bills_group_id ON bills(group_id);
```

### 📊 BillSplit
```sql
CREATE TABLE bill_splits (
  id SERIAL PRIMARY KEY,
  group_id INT REFERENCES groups(id),
  user_id INT REFERENCES users(id),
  to_pay_user_id INT REFERENCES users(id),
  amount_due FLOAT NOT NULL,
  is_paid BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
CREATE INDEX idx_splits_group_id ON bill_splits(group_id);
```

---

## 🔗 Entity Relationships

- **User** can belong to many **Groups**
- **Group** can have many **Users** with specific **Permissions**
- **User** can add multiple **Bills** to a **Group**
- **Bills** are split using **BillSplits**, where `user_id` owes `to_pay_user_id`
- **AuthToken** and **OTP** are associated with **User** for auth flows

---

## 🧪 API Base Routes (REST)

| Method | Endpoint                                   | Description                     |
|--------|--------------------------------------------|---------------------------------|
| POST   | `/api/v1/users/register`                   | Register a user                 |
| POST   | `/api/v1/users/activate`                   | Activate with OTP               |
| POST   | `/api/v1/users/login`                      | Login and get access token      |
| POST   | `/api/v1/groups`                           | Create a new group              |
| PUT    | `/api/v1/groups/:group_id`                 | Update group info               |
| DELETE | `/api/v1/groups/:group_id`                 | Delete group                    |
| GET    | `/api/v1/groups`                           | List user groups                |
| POST   | `/api/v1/groups/:group_id/assign/:user_id` | Assign a user to group          |
| POST   | `/api/v1/groups/:group_id/users/:user_id/bills` | Add bill to group         |
| PUT    | `/api/v1/groups/:group_id/bills/:bill_id`  | Update bill                     |
| DELETE | `/api/v1/groups/:group_id/bills/:bill_id`  | Delete bill                     |
| POST   | `/api/v1/groups/:group_id/splits`          | Calculate bill splits           |
| PUT    | `/api/v1/groups/:group_id/splits`          | Recalculate bill splits         |

---

## 📌 Notes

- All timestamps are `UTC`
- Soft deletes used via `DeletedAt`
- Permissions are **granular per user per group**
- `BillSplit` reflects **who owes how much to whom**

---

> ✅ Ready for deployment with PostgreSQL, JWT Auth, and REST APIs
> For frontend testing, use Postman 
