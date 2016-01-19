package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/mcoffin/terraform-provider-chronos/chronos"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: chronos.Provider,
	})
}
