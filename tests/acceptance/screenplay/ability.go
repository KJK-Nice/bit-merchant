package screenplay

import (
	"fmt"

	"github.com/go-rod/rod"
)

const AbilityBrowseTheWeb = "BrowseTheWeb"

// BrowseTheWeb gives an actor the ability to browse the web using Rod.
type BrowseTheWeb struct {
	Page *rod.Page
}

// Name returns the name of the ability.
func (b *BrowseTheWeb) Name() string {
	return AbilityBrowseTheWeb
}

// UsingRod creates a new BrowseTheWeb ability.
func UsingRod(page *rod.Page) *BrowseTheWeb {
	return &BrowseTheWeb{Page: page}
}

// BrowseTheWebWith retrieves the Rod page from the actor.
func BrowseTheWebWith(actor *Actor) (*rod.Page, error) {
	ability := actor.Ability(AbilityBrowseTheWeb)
	if ability == nil {
		return nil, fmt.Errorf("actor %s does not have the ability to BrowseTheWeb", actor.Name)
	}
	return ability.(*BrowseTheWeb).Page, nil
}

