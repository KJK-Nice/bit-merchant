package postgres

import authAdapters "bitmerchant/internal/auth/adapters"

type MembershipRepository = authAdapters.PostgresMembershipRepository

var NewMembershipRepository = authAdapters.NewPostgresMembershipRepository
