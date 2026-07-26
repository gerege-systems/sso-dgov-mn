// Package main нь Government Single Sign On-ий API эхлэх цэг.
//
// Бүх суурь чадвар (танилт, RBAC, API gateway, OIDC provider, eID proxy) нь
// github.com/gerege-systems/platform-core модульд байрлана.
package main

import (
	"github.com/gerege-systems/platform-core/cmd/api/server"
	"github.com/gerege-systems/platform-core/core/constants"
	"github.com/gerege-systems/platform-core/pkg/logger"
)

func main() {
	server.ServiceName = "sso-dgov"

	app, err := server.NewApp()
	if err != nil {
		logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryServer})
	}

	if err := app.Run(); err != nil {
		logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryServer})
	}
}
