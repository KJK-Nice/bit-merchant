package http_test

import (
	"strings"
	"testing"

	commonhttp "bitmerchant/internal/common/http"

	"github.com/stretchr/testify/assert"
)

func TestFormatDatastarEvent(t *testing.T) {
	fragment := "<div id='test'>Content</div>"
	result := commonhttp.FormatDatastarEvent(fragment)
	str := string(result)

	assert.Contains(t, str, "event: datastar-patch-elements")
	assert.Contains(t, str, "data: elements <div id='test'>Content</div>")
	assert.True(t, strings.HasSuffix(str, "\n\n"))
}

func TestFormatDatastarPatch(t *testing.T) {
	fragment := "<div id='item'>Item</div>"
	selector := "#list"
	mode := "prepend"

	result := commonhttp.FormatDatastarPatch(fragment, selector, mode)
	str := string(result)

	assert.Contains(t, str, "event: datastar-patch-elements")
	assert.Contains(t, str, "data: selector #list")
	assert.Contains(t, str, "data: mode prepend")
	assert.Contains(t, str, "data: elements <div id='item'>Item</div>")
	assert.True(t, strings.HasSuffix(str, "\n\n"))
}
