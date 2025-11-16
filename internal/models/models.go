package models

import "time"

type Subscriber struct {
  ID    int
  Email string
}

type Newsletter struct {
  ID      int
  Subject string
  Body    string
  SentAt  time.Time
}