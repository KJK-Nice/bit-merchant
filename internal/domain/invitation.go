package domain

import (
	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/common"
)

type InvitationID = common.InvitationID
type Invitation = invitation.Invitation

var NewInvitation = invitation.NewInvitation
