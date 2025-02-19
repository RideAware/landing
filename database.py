import os
import psycopg2
from psycopg2 import IntegrityError
from dotenv import load_dotenv

load_dotenv()

def get_connection():
    """Return a database connection."""
    return psycopg2.connect(
        host=os.getenv("PG_HOST"),
        port=os.getenv("PG_PORT"),
        dbname=os.getenv("PG_DATABASE"),
        user=os.getenv("PG_USER"),
        password=os.getenv("PG_PASSWORD"),
        connect_timeout=10
    )

def init_db():
    conn = get_connection()
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS subscribers (
            id SERIAL PRIMARY KEY,
            email TEXT UNIQUE NOT NULL
        )
    """)
    
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS newsletters(
            id SERIAL PRIMARY KEY,
            subject TEXT NOT NULL,
            body TEXT NOT NULL,
            sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    """)
        
    conn.commit()
    cursor.close()
    conn.close()

def add_email(email):
    try:
        with get_connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute("INSERT INTO subscribers (email) VALUES (%s)", (email,))
                conn.commit()
        return True
    except IntegrityError:
        return False
    except psycopg2.OperationalError as e:
        print(f"Error: {e}")
        return False

def remove_email(email):
    try:
        with get_connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute("DELETE FROM subscribers WHERE email = %s", (email,))
                conn.commit()
                if cursor.rowcount > 0:
                    return True
                return False
    except Exception as e:
        print(f"Error removing email: {e}")
        return False