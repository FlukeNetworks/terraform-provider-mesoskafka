package main

import (
	"github.com/FlukeNetworks/terraform-provider-mesoskafka/mesoskafka"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: mesoskafka.Provider,
	})
}
