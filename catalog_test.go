package main

import (
	"path/filepath"
	"testing"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/catalog"
	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"

	"github.com/gerege-systems/sso-gerege-nexus/modules/accessreview"
	"github.com/gerege-systems/sso-gerege-nexus/modules/federation"
	"github.com/gerege-systems/sso-gerege-nexus/modules/provisioning"
	"github.com/gerege-systems/sso-gerege-nexus/modules/sessions"
)

// The bundled catalogue has to agree with the binary it ships inside.
//
// The platform refuses to start on a catalogue whose apps disagree with the
// modules compiled into it — catalogue integrity is a boot failure, not a
// warning — so a disagreement here is not a stale file, it is an instance that
// does not come up. And it is easy to arrive at without touching this file:
// bumping the core is a one-line change, and a core release that renames one of
// the platform apps this catalogue copies leaves the copy behind. The App Store
// hit exactly that on backend/v1.5.0, when io.gerege.nexus.organisation became
// "Directory" 2.0.0 and its copy still said 1.0.0.
//
// What this test can check is the half that lives in this repository: the four
// SSO modules, which it compiles and can therefore ask. The five platform apps
// it copies are checked at boot by the core, because their code is not
// importable from here — it lives in the core's internal/. See README.
func TestTheBundledCatalogueAgreesWithThisBinary(t *testing.T) {
	apps := loadBundledCatalogue(t)

	// Registering the modules is what makes them askable. A nil platform is
	// enough: nothing below calls a method that touches one.
	p := nexus.NewPlatform(nil, nil)
	accessreview.New(p)
	federation.New(p)
	provisioning.New(p)
	sessions.New(p)

	for _, app := range apps {
		module, compiled := nexus.Get(app.ID)
		if !compiled {
			// A platform app from the core. A legitimate entry this test
			// cannot speak for.
			continue
		}
		if module.Name() != app.Manifest.Name {
			t.Errorf("%s is compiled as %q and its manifest says %q",
				app.ID, module.Name(), app.Manifest.Name)
		}
		if module.Version() != app.Version {
			t.Errorf("%s is compiled at %s and the catalogue says %s",
				app.ID, module.Version(), app.Version)
		}
		if module.Version() != app.Manifest.Version {
			t.Errorf("%s is compiled at %s and its manifest says %s",
				app.ID, module.Version(), app.Manifest.Version)
		}
	}
}

// Every module this product owns is in the catalogue it ships.
//
// The test above runs from the catalogue inwards and is silent about a module
// that is compiled in and listed nowhere — which is not a boot failure but is a
// screen nobody can install, and the failure mode of adding a module and
// forgetting the four catalogue files that go with it.
func TestEveryModuleThisRepositoryOwnsIsInTheCatalogue(t *testing.T) {
	apps := loadBundledCatalogue(t)
	listed := make(map[string]bool, len(apps))
	for _, app := range apps {
		listed[app.ID] = true
	}

	for _, id := range []string{
		accessreview.ID, federation.ID, provisioning.ID, sessions.ID,
	} {
		if !listed[id] {
			t.Errorf("%s is compiled into this binary and is not in catalog/apps.json", id)
		}
	}
}

// Every entry declares a visibility the contract knows.
//
// An unknown third value is refused rather than read as either: taken as public
// it would turn a typo into a publication, taken as private it would hide an
// app for a reason nobody could see.
func TestEveryBundledAppDeclaresAKnownVisibility(t *testing.T) {
	for _, app := range loadBundledCatalogue(t) {
		switch app.Visibility {
		case "", catalog.VisibilityPublic, catalog.VisibilityPrivate:
		default:
			t.Errorf("%s declares visibility %q; expected %q or %q",
				app.ID, app.Visibility, catalog.VisibilityPublic, catalog.VisibilityPrivate)
		}
	}
}

// loadBundledCatalogue reads catalog/apps.json and the manifests beside it.
//
// The platform version is left empty on purpose. It is stamped into the binary
// at build time by -ldflags and lives in an internal package, so a test cannot
// read the real one; empty skips the compatibility constraint and leaves
// everything this file is about — names, versions, manifests, visibility —
// checked.
func loadBundledCatalogue(t *testing.T) []catalog.CatalogApp {
	t.Helper()
	apps, err := catalog.LoadFile(filepath.Join("catalog", "apps.json"), "")
	if err != nil {
		t.Fatalf("the bundled catalogue does not load: %v", err)
	}
	if len(apps) == 0 {
		t.Fatal("the bundled catalogue is empty")
	}
	return apps
}
