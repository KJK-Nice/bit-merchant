package memory

import authAdapters "bitmerchant/internal/auth/adapters"

type MemoryUserRepository = authAdapters.MemoryUserRepository

var NewMemoryUserRepository = authAdapters.NewMemoryUserRepository
