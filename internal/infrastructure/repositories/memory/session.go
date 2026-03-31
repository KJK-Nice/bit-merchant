package memory

import authAdapters "bitmerchant/internal/auth/adapters"

type MemorySessionRepository = authAdapters.MemorySessionRepository

var NewMemorySessionRepository = authAdapters.NewMemorySessionRepository
