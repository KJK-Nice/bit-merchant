package dashboard

import dashQuery "bitmerchant/internal/dashboard/app/query"

type DateRange = dashQuery.DateRange
type DashboardStats = dashQuery.DashboardStats
type GetDashboardStatsUseCase = dashQuery.GetDashboardStatsUseCase

var NewGetDashboardStatsUseCase = dashQuery.NewGetDashboardStatsUseCase

const (
	DateRangeToday = dashQuery.DateRangeToday
	DateRangeWeek  = dashQuery.DateRangeWeek
	DateRangeMonth = dashQuery.DateRangeMonth
)
