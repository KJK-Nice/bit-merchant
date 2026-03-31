package postgres

import authAdapters "bitmerchant/internal/auth/adapters"

type UserRepository = authAdapters.PostgresUserRepository

var NewUserRepository = authAdapters.NewPostgresUserRepository
