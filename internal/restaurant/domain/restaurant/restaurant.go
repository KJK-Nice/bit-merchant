package restaurant

import (
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
)

const (
	MinTableCount = 1
	MaxTableCount = 200
	// DefaultTaxRate applies to new restaurants when none is specified.
	DefaultTaxRate = 0.08

	// Kitchen escalation thresholds (in minutes). An order is "warning" once its
	// age crosses the warning threshold and "overdue" once it crosses the overdue
	// threshold; below the warning threshold it is nominal. Defaults match the
	// values the kitchen board used before per-restaurant configuration (#76).
	DefaultKitchenWarningMinutes = 8
	DefaultKitchenOverdueMinutes = 12
	MinKitchenThresholdMinutes   = 1
	MaxKitchenThresholdMinutes   = 120
)

var (
	ErrInvalidTableCount        = errors.New("invalid table count")
	ErrInvalidTaxRate           = errors.New("invalid tax rate")
	ErrInvalidKitchenThresholds = errors.New("kitchen thresholds must satisfy 1 <= warning < overdue <= 120 minutes")
)

// Restaurant represents a single restaurant tenant.
type Restaurant struct {
	ID             common.RestaurantID
	Name           string
	BaseCurrency   money.Currency
	TaxRate        float64 // 0.08 = 8%
	TableCount     int
	IsOpen         bool
	ClosedMessage  string
	ReopeningHours string
	// KitchenWarningMinutes / KitchenOverdueMinutes drive the kitchen board's
	// escalation tiers (nominal / warning / overdue). Configurable per restaurant.
	KitchenWarningMinutes int
	KitchenOverdueMinutes int
	// PausedUntil is non-nil when the owner has applied a quick-pause
	// (rush). The restaurant auto-resumes once now passes this timestamp;
	// readers should call AcceptingOrdersAt to apply that lazily.
	PausedUntil *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewRestaurant creates a new Restaurant with validation. The currency
// defaults to USD; use NewRestaurantWithCurrency to pick THB or SAT.
func NewRestaurant(id common.RestaurantID, name string) (*Restaurant, error) {
	return NewRestaurantWithCurrency(id, name, money.USD)
}

// NewRestaurantWithCurrency creates a Restaurant priced in the given base
// currency. All menu items, orders, and payments inherit this currency.
func NewRestaurantWithCurrency(id common.RestaurantID, name string, currency money.Currency) (*Restaurant, error) {
	if err := ValidateRestaurantName(name); err != nil {
		return nil, err
	}
	if currency.IsZero() {
		currency = money.USD
	}

	now := time.Now()
	return &Restaurant{
		ID:                    id,
		Name:                  name,
		BaseCurrency:          currency,
		TaxRate:               DefaultTaxRate,
		TableCount:            MinTableCount,
		IsOpen:                true,
		KitchenWarningMinutes: DefaultKitchenWarningMinutes,
		KitchenOverdueMinutes: DefaultKitchenOverdueMinutes,
		CreatedAt:             now,
		UpdatedAt:             now,
	}, nil
}

// ValidateKitchenThresholds enforces 1 <= warning < overdue <= 120 (minutes).
func ValidateKitchenThresholds(warningMinutes, overdueMinutes int) error {
	if warningMinutes < MinKitchenThresholdMinutes || overdueMinutes > MaxKitchenThresholdMinutes {
		return ErrInvalidKitchenThresholds
	}
	if warningMinutes >= overdueMinutes {
		return ErrInvalidKitchenThresholds
	}
	return nil
}

// SetKitchenThresholds validates and applies new escalation thresholds.
func (r *Restaurant) SetKitchenThresholds(warningMinutes, overdueMinutes int) error {
	if err := ValidateKitchenThresholds(warningMinutes, overdueMinutes); err != nil {
		return err
	}
	r.KitchenWarningMinutes = warningMinutes
	r.KitchenOverdueMinutes = overdueMinutes
	r.UpdatedAt = time.Now()
	return nil
}

// EffectiveKitchenWarningMinutes returns the configured warning threshold,
// falling back to the default for legacy rows that stored zero.
func (r *Restaurant) EffectiveKitchenWarningMinutes() int {
	if r.KitchenWarningMinutes <= 0 {
		return DefaultKitchenWarningMinutes
	}
	return r.KitchenWarningMinutes
}

// EffectiveKitchenOverdueMinutes returns the configured overdue threshold,
// falling back to the default for legacy rows that stored zero.
func (r *Restaurant) EffectiveKitchenOverdueMinutes() int {
	if r.KitchenOverdueMinutes <= 0 {
		return DefaultKitchenOverdueMinutes
	}
	return r.KitchenOverdueMinutes
}

// ValidateTaxRate rejects negative rates and rates >= 1 (i.e., >= 100%).
func ValidateTaxRate(rate float64) error {
	if rate < 0 || rate >= 1 {
		return ErrInvalidTaxRate
	}
	return nil
}

func ValidateRestaurantName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return errors.New("restaurant name must be between 1 and 100 characters")
	}
	return nil
}

func ValidateTableCount(n int) error {
	if n < MinTableCount || n > MaxTableCount {
		return ErrInvalidTableCount
	}
	return nil
}

func ValidateDescription(description string) error {
	if len(description) > 500 {
		return errors.New("description must be <= 500 characters")
	}
	return nil
}

// Open marks the restaurant as open, clearing closed messaging and any
// active quick-pause window.
func (r *Restaurant) Open() {
	r.IsOpen = true
	r.ClosedMessage = ""
	r.ReopeningHours = ""
	r.PausedUntil = nil
	r.UpdatedAt = time.Now()
}

// Close marks the restaurant as closed with optional messaging. Any active
// quick-pause is cleared — closing is a stronger signal than pausing.
func (r *Restaurant) Close(closedMessage, reopeningHours string) {
	r.IsOpen = false
	r.ClosedMessage = closedMessage
	r.ReopeningHours = reopeningHours
	r.PausedUntil = nil
	r.UpdatedAt = time.Now()
}

// MaxPauseDuration caps quick-pause windows; longer windows should use the
// regular close flow with messaging.
const MaxPauseDuration = 4 * time.Hour

// ErrInvalidPauseDuration indicates the requested pause window is non-positive
// or beyond MaxPauseDuration.
var ErrInvalidPauseDuration = errors.New("pause duration must be 1m..4h")

// Pause keeps IsOpen=true but suppresses new orders until now+duration.
// The restaurant auto-resumes lazily via AcceptingOrdersAt.
func (r *Restaurant) Pause(now time.Time, duration time.Duration) error {
	if duration <= 0 || duration > MaxPauseDuration {
		return ErrInvalidPauseDuration
	}
	until := now.Add(duration)
	r.PausedUntil = &until
	r.UpdatedAt = now
	return nil
}

// Resume clears any pending pause window without changing IsOpen.
func (r *Restaurant) Resume() {
	r.PausedUntil = nil
	r.UpdatedAt = time.Now()
}

// AcceptingOrdersAt reports whether the restaurant is currently taking orders,
// applying any active quick-pause window relative to the given instant.
func (r *Restaurant) AcceptingOrdersAt(now time.Time) bool {
	if !r.IsOpen {
		return false
	}
	if r.PausedUntil != nil && now.Before(*r.PausedUntil) {
		return false
	}
	return true
}

// IsPausedAt reports whether the restaurant is mid-pause at the given instant.
func (r *Restaurant) IsPausedAt(now time.Time) bool {
	return r.IsOpen && r.PausedUntil != nil && now.Before(*r.PausedUntil)
}

// UpdateStatus updates restaurant open/closed status (kept for backward compatibility).
func (r *Restaurant) UpdateStatus(isOpen bool, closedMessage, reopeningHours string) {
	r.IsOpen = isOpen
	r.ClosedMessage = closedMessage
	r.ReopeningHours = reopeningHours
	r.UpdatedAt = time.Now()
}
