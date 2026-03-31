package main

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

func seedData(repos repositories) {
	restaurantID := common.RestaurantID("restaurant_1")
	restaurantObj, _ := restaurant.NewRestaurant(restaurantID, "BitMerchant Cafe")
	_ = repos.Restaurant.Save(restaurantObj)

	cat1, _ := menu.NewMenuCategory("cat_1", restaurantID, "Appetizers", 1)
	cat2, _ := menu.NewMenuCategory("cat_2", restaurantID, "Mains", 2)
	cat3, _ := menu.NewMenuCategory("cat_3", restaurantID, "Drinks", 3)
	_ = repos.MenuCategory.Save(cat1)
	_ = repos.MenuCategory.Save(cat2)
	_ = repos.MenuCategory.Save(cat3)

	item1, _ := menu.NewMenuItem("item_1", "cat_1", restaurantID, "Bruschetta", 8.50)
	_ = item1.SetDescription("Toasted bread with tomatoes and basil")
	_ = repos.MenuItem.Save(item1)

	item2, _ := menu.NewMenuItem("item_2", "cat_2", restaurantID, "Bitcoin Burger", 15.00)
	_ = item2.SetDescription("Premium beef patty with cheese")
	_ = repos.MenuItem.Save(item2)

	item3, _ := menu.NewMenuItem("item_3", "cat_3", restaurantID, "Satoshi Soda", 3.00)
	_ = repos.MenuItem.Save(item3)
}
