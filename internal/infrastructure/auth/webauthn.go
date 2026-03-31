package auth

import authAdapters "bitmerchant/internal/auth/adapters"

type WebAuthnService = authAdapters.WebAuthnService

var NewWebAuthnService = authAdapters.NewWebAuthnService
