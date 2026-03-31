package postgres

import authAdapters "bitmerchant/internal/auth/adapters"

type InvitationRepository = authAdapters.PostgresInvitationRepository

var NewInvitationRepository = authAdapters.NewPostgresInvitationRepository
