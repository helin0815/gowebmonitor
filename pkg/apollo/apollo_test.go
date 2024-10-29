package apollo

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestApolloGet(t *testing.T) {
	cc := NewApollo("servicemesh-api", "dev")
	assert.NoError(t, cc.Run())
	assert.NotEmpty(t, viper.GetString("idaas.url"))
}
