package healthcheck

import (
	"context"
	healthcheckv1 "go.admiral.io/admiral/api/healthcheck/v1"
	"go.admiral.io/admiral/internal/endpoint/endpointtest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap/zaptest"
)

func TestModule(t *testing.T) {
	log := zaptest.NewLogger(t)
	scope := tally.NewTestScope("", nil)

	m, err := New(nil, log, scope)
	assert.NoError(t, err)

	r := endpointtest.NewRegisterChecker()
	assert.NoError(t, m.Register(r))
	assert.NoError(t, r.HasAPI("admiral.healthcheck.v1.HealthcheckAPI"))
	assert.True(t, r.JSONRegistered())
}

func TestAPI(t *testing.T) {
	endp := api{}
	resp, err := endp.Healthcheck(context.Background(), &healthcheckv1.HealthcheckRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
