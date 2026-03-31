package memory

import authAdapters "bitmerchant/internal/auth/adapters"

type MemoryMembershipRepository = authAdapters.MemoryMembershipRepository

var NewMemoryMembershipRepository = authAdapters.NewMemoryMembershipRepository
