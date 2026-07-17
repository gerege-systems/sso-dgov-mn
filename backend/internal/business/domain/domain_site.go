// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package domain

import (
	"regexp"
	"time"
)

// SiteAppearance нь сайтын нийтийн харагдацын default — админ тохируулж, бүх
// зочин үүгээр эхэлнэ. accent нь preset нэр ЭСВЭЛ '#rrggbb' custom hex.
type SiteAppearance struct {
	Accent    string     `db:"accent"`
	Font      string     `db:"font"`
	Style     string     `db:"style"`
	Theme     string     `db:"theme"`
	LandingBg string     `db:"landing_bg"`
	UpdatedAt *time.Time `db:"updated_at"`
}

// Зөвшөөрөгдсөн утгууд — frontend-ийн preset жагсаалттай нэг мөр байх ёстой
// (globals.css html[data-*], preferences.ts).
var (
	SiteAccentPresets = map[string]bool{"cobalt": true, "teal": true, "violet": true, "emerald": true, "amber": true}
	SiteFonts         = map[string]bool{"inter": true, "serif": true, "system": true}
	SiteStyles        = map[string]bool{"comfortable": true, "compact": true}
	SiteThemes        = map[string]bool{"light": true, "dark": true, "system": true}
)

// siteHexRe нь custom accent/landing-bg-ийн '#rrggbb' хэлбэрийг шалгана
// (3-оронтой хэлбэр зөвшөөрөхгүй — frontend нь 6-оронтойг л илгээнэ).
var siteHexRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// DefaultLandingBg нь landing navy дэвсгэрийн built-in default (globals.css-ийн
// --lp-navy oklch(24% .055 260)-ийн hex дүйцэл).
const DefaultLandingBg = "#0f1f39"

// DefaultSiteAppearance нь seed/fallback утга (repo уншиж чадаагүй үед ч ашиглана).
func DefaultSiteAppearance() SiteAppearance {
	return SiteAppearance{Accent: "cobalt", Font: "inter", Style: "comfortable", Theme: "light", LandingBg: DefaultLandingBg}
}

// ValidSiteAccent нь preset нэр эсвэл '#rrggbb' hex мөнийг шалгана.
func ValidSiteAccent(accent string) bool {
	return SiteAccentPresets[accent] || siteHexRe.MatchString(accent)
}

// ValidLandingBg нь landing дэвсгэрийн '#rrggbb' hex мөнийг шалгана.
func ValidLandingBg(bg string) bool {
	return siteHexRe.MatchString(bg)
}
