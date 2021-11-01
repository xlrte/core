package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlrte/core/pkg/api"
)

func Test_Configures_CloudRun_With_Network(t *testing.T) {
	binding := privateNetworkBinding{
		identity: api.ResourceIdentity{ID: "network", Type: "private_network"},
	}
	cloudRun := cloudRunConfig{}

	err := binding.ConfigureResource(&cloudRun)
	assert.NoError(t, err)
	assert.True(t, cloudRun.HasServerlessNetwork)
	assert.Equal(t, cloudRun.ServerlessNetworkLink, "module.private_network-network.serverless_connector")
	assert.Equal(t, cloudRun.DependsOn, []string{"module.private_network-network.serverless_connector"})
}

func Test_Configures_CloudSql_With_Network(t *testing.T) {
	binding := privateNetworkBinding{
		identity: api.ResourceIdentity{ID: "network", Type: "private_network"},
	}
	sql := cloudSql{}

	err := binding.ConfigureResource(&sql)
	assert.NoError(t, err)
	assert.Equal(t, sql.NetworkLink, "module.private_network-network.network_self_link")

}
