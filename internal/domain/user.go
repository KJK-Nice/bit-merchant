package domain

import (
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"
)

type UserID = common.UserID
type User = user.User

var NewUser = user.NewUser
