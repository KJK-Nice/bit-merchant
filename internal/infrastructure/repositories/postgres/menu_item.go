package postgres

import menuAdapters "bitmerchant/internal/menu/adapters"

type MenuItemRepository = menuAdapters.PostgresItemRepository

var NewMenuItemRepository = menuAdapters.NewPostgresItemRepository
