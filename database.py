import sqlite3
import os
from dotenv import load_dotenv

load_dotenv()

DATABASE_URL = os.getenv("DATABASE_FILE")

def init_db():
    conn = sqlite3.connect(DATABASE_URL)
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS subscribers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT UNIQUE NOT NULL
        )
    """)
    conn.commit()
    conn.close()

def add_email(email):
    try:
        with sqlite3.connect(DATABASE_URL, timeout=10) as conn:
            cursor = conn.cursor()
            cursor.execute("INSERT INTO subscribers (email) VALUES (?)", (email,))
            conn.commit()
        return True
    except sqlite3.IntegrityError:
        return False
    except sqlite3.OperationalError as e:
        print(f"Operational Error: {e}")
        return False

def remove_email(email):
    try:
        with sqlite3.connect(DATABASE_URL, timeout=10) as conn:
            cursor = conn.cursor()
            cursor.execute("DELETE FROM subscribers WHERE email = ?", (email,))
            conn.commit()
            if cursor.rowcount > 0:
                return True
            return False
    except Exception as e:
        print(f"Error removing email: {e}")
        return False