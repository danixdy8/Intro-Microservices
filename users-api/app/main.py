import hashlib
import sqlite3
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, HTTPException, status
from pydantic import BaseModel, Field

from app.database import DB_PATH, get_connection, init_db


@asynccontextmanager
async def lifespan(_: FastAPI):
    init_db()
    yield


app = FastAPI(title="Users API", version="1.0.0", lifespan=lifespan)


class UserCreate(BaseModel):
    name: str = Field(..., min_length=2)
    email: str = Field(..., min_length=4)
    password: str = Field(..., min_length=4)
    role: str = Field(default="user")


class UserUpdate(BaseModel):
    name: Optional[str] = Field(default=None, min_length=2)
    email: Optional[str] = Field(default=None, min_length=4)
    password: Optional[str] = Field(default=None, min_length=4)
    role: Optional[str] = None


def hash_password(raw_password: str) -> str:
    return hashlib.sha256(f"users_api_salt::{raw_password}".encode("utf-8")).hexdigest()


def serialize_user(row: sqlite3.Row) -> dict:
    return {
        "id": row["id"],
        "name": row["name"],
        "email": row["email"],
        "role": row["role"],
        "created_at": row["created_at"],
        "updated_at": row["updated_at"],
    }


@app.get("/health")
def health():
    with get_connection() as conn:
        count = conn.execute("SELECT COUNT(*) AS total FROM users").fetchone()["total"]
    return {
        "service": "users-api",
        "language": "Python / FastAPI",
        "status": "UP",
        "database": DB_PATH,
        "records": count,
    }


@app.post("/users", status_code=status.HTTP_201_CREATED)
def create_user(payload: UserCreate):
    try:
        with get_connection() as conn:
            cursor = conn.execute(
                """
                INSERT INTO users (name, email, password_hash, role)
                VALUES (?, ?, ?, ?)
                """,
                (
                    payload.name.strip(),
                    payload.email.strip().lower(),
                    hash_password(payload.password),
                    payload.role.strip(),
                ),
            )
            conn.commit()
            row = conn.execute("SELECT * FROM users WHERE id = ?", (cursor.lastrowid,)).fetchone()
            return {"message": "Usuario registrado correctamente", "data": serialize_user(row)}
    except sqlite3.IntegrityError:
        raise HTTPException(status_code=409, detail="Ya existe un usuario con ese correo electrónico")


@app.get("/users")
def list_users():
    with get_connection() as conn:
        rows = conn.execute("SELECT * FROM users ORDER BY id ASC").fetchall()
    return {"total": len(rows), "data": [serialize_user(r) for r in rows]}


@app.get("/users/{user_id}")
def get_user(user_id: int):
    with get_connection() as conn:
        row = conn.execute("SELECT * FROM users WHERE id = ?", (user_id,)).fetchone()
    if not row:
        raise HTTPException(status_code=404, detail=f"Usuario con ID {user_id} no encontrado")
    return {"data": serialize_user(row)}


@app.put("/users/{user_id}")
def update_user(user_id: int, payload: UserUpdate):
    with get_connection() as conn:
        existing = conn.execute("SELECT * FROM users WHERE id = ?", (user_id,)).fetchone()
        if not existing:
            raise HTTPException(status_code=404, detail=f"Usuario con ID {user_id} no encontrado")

        new_name = payload.name.strip() if payload.name is not None else existing["name"]
        new_email = payload.email.strip().lower() if payload.email is not None else existing["email"]
        new_role = payload.role.strip() if payload.role is not None else existing["role"]
        new_hash = hash_password(payload.password) if payload.password is not None else existing["password_hash"]

        try:
            conn.execute(
                """
                UPDATE users
                SET name = ?, email = ?, password_hash = ?, role = ?, updated_at = datetime('now')
                WHERE id = ?
                """,
                (new_name, new_email, new_hash, new_role, user_id),
            )
            conn.commit()
            updated = conn.execute("SELECT * FROM users WHERE id = ?", (user_id,)).fetchone()
            return {"message": "Usuario actualizado correctamente", "data": serialize_user(updated)}
        except sqlite3.IntegrityError:
            raise HTTPException(status_code=409, detail="El correo electrónico ya pertenece a otro usuario")


@app.delete("/users/{user_id}")
def delete_user(user_id: int):
    with get_connection() as conn:
        cursor = conn.execute("DELETE FROM users WHERE id = ?", (user_id,))
        conn.commit()
    if cursor.rowcount == 0:
        raise HTTPException(status_code=404, detail=f"Usuario con ID {user_id} no encontrado")
    return {"message": f"Usuario {user_id} eliminado correctamente", "deleted_id": user_id}
