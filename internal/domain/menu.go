package domain

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type CategoryID = common.CategoryID
type ItemID = common.ItemID
type MenuCategory = menu.MenuCategory
type MenuItem = menu.MenuItem

var NewMenuCategory = menu.NewMenuCategory
var NewMenuItem = menu.NewMenuItem
var ValidateCategoryName = menu.ValidateCategoryName
var ValidateItemName = menu.ValidateItemName
var ValidatePrice = menu.ValidatePrice
var ValidateDescription = menu.ValidateDescription
