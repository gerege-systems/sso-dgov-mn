/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 *
 * The whole of this product's wiring.
 *
 * There is no platform code here and there is not meant to be: sign-in,
 * tenants, the database and its isolation, OAuth2, the app gate, the menu and
 * the HTTP server are the core's, taken as a dependency by tag. What this
 * repository adds is four modules and the line that registers them.
 *
 * This repository used to be a fork of the platform — 698 files apart from it —
 * and everything above is what that fork was carrying a second copy of. See the
 * README for what replaced it and why.
 *
 * If this file ever grows business logic, it belongs in a module instead — see
 * CONTRIBUTING in the core repository, rule 3.
 */

package main

import (
	"log/slog"
	"os"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/platform"

	"github.com/gerege-systems/sso-gerege-nexus/modules/accessreview"
	"github.com/gerege-systems/sso-gerege-nexus/modules/federation"
	"github.com/gerege-systems/sso-gerege-nexus/modules/provisioning"
	"github.com/gerege-systems/sso-gerege-nexus/modules/sessions"
)

func main() {
	err := platform.Run(platform.Options{
		// Alphabetical, because for once the order carries no meaning: no
		// module here constructs another, and none reads a table another one
		// owns. Ordering them by anything else would imply a dependency that
		// the next person would then have to go and look for.
		Modules: func(p nexus.Platform) {
			accessreview.New(p)
			federation.New(p)
			provisioning.New(p)
			sessions.New(p)
		},
	})
	if err != nil {
		slog.Error("Gerege SSO stopped", "error", err)
		os.Exit(1)
	}
}
