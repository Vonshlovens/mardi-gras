package ui

import (
	"fmt"
	"image/color"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

func glamourStr(s string) *string { return &s }
func glamourBool(b bool) *bool    { return &b }

func glamourHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

// MardiGrasGlamourStyle returns the active Mardi Gras/Tuxedo palette as a
// glamour StyleConfig. The historical name is kept because Detail uses this
// exported function; its output is now fully theme-derived rather than a
// permanently Mardi Gras-coloured markdown renderer.
func MardiGrasGlamourStyle() ansi.StyleConfig {
	theme := CurrentTheme()
	if theme.Terminal {
		// Terminal must respect the user's terminal palette. Glamour's ASCII
		// config intentionally emits no hard-coded foreground/background ANSI
		// colours, unlike its dark preset.
		return styles.ASCIIStyleConfig
	}

	c := styles.DarkStyleConfig

	// Apply the active palette to all prominent prose elements. This overrides
	// DarkStyleConfig's old hard-coded colours on each call without mutating the
	// package-level base config.
	c.Document.Color = glamourStr(glamourHex(White))
	c.H1.Color = glamourStr(glamourHex(BrightGold))
	c.H1.BackgroundColor = glamourStr(glamourHex(Purple))
	c.H1.Bold = glamourBool(true)
	c.Heading.Color = glamourStr(glamourHex(BrightPurple))
	c.Heading.Bold = glamourBool(true)
	c.H6.Color = glamourStr(glamourHex(BrightPurple))
	c.H6.Bold = glamourBool(true)
	c.Emph.Color = glamourStr(glamourHex(Gold))
	c.Strong.Color = glamourStr(glamourHex(BrightGold))
	c.Strong.Bold = glamourBool(true)
	c.Link.Color = glamourStr(glamourHex(Gold))
	c.Link.Underline = glamourBool(true)
	c.LinkText.Color = glamourStr(glamourHex(Gold))
	c.Image.Color = glamourStr(glamourHex(BrightPurple))
	c.ImageText.Color = glamourStr(glamourHex(Muted))
	c.HorizontalRule.Color = glamourStr(glamourHex(Dim))
	c.Code.Color = glamourStr(glamourHex(BrightGreen))
	c.Code.BackgroundColor = glamourStr(glamourHex(Panel))
	c.CodeBlock.Color = glamourStr(glamourHex(Light))
	// Syntax colours in the cloned dark config are not theme-aware. Leaving
	// them unset is preferable to leaking an unrelated palette into a selected
	// Tuxedo theme; code blocks retain their structure and base foreground.
	c.CodeBlock.Chroma = nil
	c.Item.Color = glamourStr(glamourHex(BrightPurple))
	c.Enumeration.Color = glamourStr(glamourHex(BrightPurple))

	return c
}
