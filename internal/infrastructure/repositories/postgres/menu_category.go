package postgres

import menuAdapters "bitmerchant/internal/menu/adapters"

type MenuCategoryRepository = menuAdapters.PostgresCategoryRepository

var NewMenuCategoryRepository = menuAdapters.NewPostgresCategoryRepository
