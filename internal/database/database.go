package database

import (
  "context"
  "fmt"
  "time"

  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgxpool"
  "landing/internal/config"
  "landing/internal/models"
)

type DB struct {
  pool *pgxpool.Pool
}

func New(cfg *config.Config) (*DB, error) {
  // Use proper pgx connection config instead of URL parsing
  connConfig, err := pgxpool.ParseConfig("")
  if err != nil {
    return nil, fmt.Errorf("failed to parse config: %w", err)
  }

  connConfig.ConnConfig.Host = cfg.DBHost
  connConfig.ConnConfig.Port = 5432
  connConfig.ConnConfig.Database = cfg.DBName
  connConfig.ConnConfig.User = cfg.DBUser
  connConfig.ConnConfig.Password = cfg.DBPass

  ctx, cancel := context.WithTimeout(
    context.Background(),
    10*time.Second,
  )
  defer cancel()

  pool, err := pgxpool.NewWithConfig(ctx, connConfig)
  if err != nil {
    return nil, fmt.Errorf("failed to create pool: %w", err)
  }

  if err := pool.Ping(ctx); err != nil {
    return nil, fmt.Errorf("failed to ping database: %w", err)
  }

  return &DB{pool: pool}, nil
}

func (db *DB) InitDB(ctx context.Context) error {
  queries := []string{
    `CREATE TABLE IF NOT EXISTS subscribers (
      id SERIAL PRIMARY KEY,
      email TEXT UNIQUE NOT NULL
    )`,
    `CREATE TABLE IF NOT EXISTS newsletters (
      id SERIAL PRIMARY KEY,
      subject TEXT NOT NULL,
      body TEXT NOT NULL,
      sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`,
  }

  for _, query := range queries {
    if _, err := db.pool.Exec(ctx, query); err != nil {
      return fmt.Errorf("failed to execute query: %w", err)
    }
  }

  return nil
}

func (db *DB) AddSubscriber(
  ctx context.Context,
  email string,
) error {
  _, err := db.pool.Exec(
    ctx,
    "INSERT INTO subscribers (email) VALUES ($1)",
    email,
  )
  if err != nil {
    if err.Error() == "ERROR: duplicate key value" {
      return fmt.Errorf("email already exists")
    }
    return err
  }
  return nil
}

func (db *DB) RemoveSubscriber(
  ctx context.Context,
  email string,
) error {
  result, err := db.pool.Exec(
    ctx,
    "DELETE FROM subscribers WHERE email = $1",
    email,
  )
  if err != nil {
    return err
  }

  if result.RowsAffected() == 0 {
    return fmt.Errorf("email not found")
  }

  return nil
}

func (db *DB) GetNewsletters(
  ctx context.Context,
) ([]models.Newsletter, error) {
  rows, err := db.pool.Query(
    ctx,
    "SELECT id, subject, body, sent_at FROM newsletters "+
      "ORDER BY sent_at DESC",
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var newsletters []models.Newsletter
  for rows.Next() {
    var n models.Newsletter
    err := rows.Scan(&n.ID, &n.Subject, &n.Body, &n.SentAt)
    if err != nil {
      return nil, err
    }
    newsletters = append(newsletters, n)
  }

  return newsletters, rows.Err()
}

func (db *DB) GetNewsletter(
  ctx context.Context,
  id int,
) (*models.Newsletter, error) {
  var n models.Newsletter
  err := db.pool.QueryRow(
    ctx,
    "SELECT id, subject, body, sent_at FROM newsletters "+
      "WHERE id = $1",
    id,
  ).Scan(&n.ID, &n.Subject, &n.Body, &n.SentAt)

  if err == pgx.ErrNoRows {
    return nil, fmt.Errorf("newsletter not found")
  }
  if err != nil {
    return nil, err
  }

  return &n, nil
}

func (db *DB) Close(ctx context.Context) {
  db.pool.Close()
}