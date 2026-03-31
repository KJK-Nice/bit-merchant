package postgres

import authAdapters "bitmerchant/internal/auth/adapters"

type SessionRepository = authAdapters.PostgresSessionRepository

var NewSessionRepository = authAdapters.NewPostgresSessionRepository
