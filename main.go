package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name deeplink

// Set at release with: -ldflags="-X main.version=vX.Y.Z"
var version = "0.1.0"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debugger support (e.g. delve)")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/yinebebt/deeplink",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), New(version), opts); err != nil {
		log.Fatal(err)
	}
}
