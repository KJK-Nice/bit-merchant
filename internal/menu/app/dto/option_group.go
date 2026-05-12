// Package dto contains data-transfer types shared by HTTP handlers and
// persistence adapters. Keeping a single JSON shape for option groups avoids
// drift between the form payload, the JSONB column, and integration fixtures.
package dto

import "bitmerchant/internal/menu/domain/menu"

// OptionDTO is the JSON shape of a single option within a group.
type OptionDTO struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	PriceDelta float64 `json:"price_delta"`
}

// OptionGroupDTO is the JSON shape persisted in menu_items.option_groups and
// posted by the admin editor form via the hidden option_groups_json field.
type OptionGroupDTO struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Required        bool        `json:"required"`
	MinSelections   int         `json:"min_selections"`
	MaxSelections   int         `json:"max_selections"`
	DefaultOptionID *string     `json:"default_option_id,omitempty"`
	Options         []OptionDTO `json:"options"`
}

// ToDomain converts a DTO slice into the domain type.
func ToDomain(in []OptionGroupDTO) []menu.OptionGroup {
	if len(in) == 0 {
		return nil
	}
	out := make([]menu.OptionGroup, len(in))
	for i, g := range in {
		opts := make([]menu.Option, len(g.Options))
		for j, o := range g.Options {
			opts[j] = menu.Option{ID: o.ID, Name: o.Name, PriceDelta: o.PriceDelta}
		}
		out[i] = menu.OptionGroup{
			ID:              g.ID,
			Name:            g.Name,
			Required:        g.Required,
			MinSelections:   g.MinSelections,
			MaxSelections:   g.MaxSelections,
			DefaultOptionID: g.DefaultOptionID,
			Options:         opts,
		}
	}
	return out
}

// FromDomain converts domain option groups into DTOs.
func FromDomain(in []menu.OptionGroup) []OptionGroupDTO {
	if len(in) == 0 {
		return nil
	}
	out := make([]OptionGroupDTO, len(in))
	for i, g := range in {
		opts := make([]OptionDTO, len(g.Options))
		for j, o := range g.Options {
			opts[j] = OptionDTO{ID: o.ID, Name: o.Name, PriceDelta: o.PriceDelta}
		}
		out[i] = OptionGroupDTO{
			ID:              g.ID,
			Name:            g.Name,
			Required:        g.Required,
			MinSelections:   g.MinSelections,
			MaxSelections:   g.MaxSelections,
			DefaultOptionID: g.DefaultOptionID,
			Options:         opts,
		}
	}
	return out
}
