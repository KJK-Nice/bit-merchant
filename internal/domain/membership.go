package domain

import (
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
)

type MembershipID = common.MembershipID
type MemberRole = common.MemberRole
type Membership = membership.Membership

var NewMembership = membership.NewMembership

const (
	RoleOwner        = common.RoleOwner
	RoleKitchenStaff = common.RoleKitchenStaff
	RoleCustomer     = common.RoleCustomer
)
