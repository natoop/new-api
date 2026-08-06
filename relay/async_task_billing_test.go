package relay

import (
	"testing"

	"github.com/QuantumNous/new-api/pkg/asynctaskbilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type frozenAsyncBillingEstimator struct {
	calls int
}

func (e *frozenAsyncBillingEstimator) EstimateAsyncTaskBilling(_ *gin.Context, _ *relaycommon.RelayInfo) (asynctaskbilling.Rule, map[string]float64, error) {
	e.calls++
	return asynctaskbilling.Rule{}, nil, nil
}

func TestPrepareAsyncTaskBillingReusesFrozenSnapshot(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	estimator := &frozenAsyncBillingEstimator{}
	info := &relaycommon.RelayInfo{
		OriginModelName: "zzdh-retry-model",
		AsyncTaskBillingSnapshot: &asynctaskbilling.Snapshot{
			BillingMode:   asynctaskbilling.BillingMode,
			GroupRatio:    1.5,
			ReservedQuota: 321,
		},
	}

	used, taskErr := prepareAsyncTaskBilling(ctx, info, estimator)
	require.Nil(t, taskErr)
	assert.True(t, used)
	assert.Zero(t, estimator.calls)
	assert.Equal(t, 321, info.PriceData.Quota)
	assert.Equal(t, 1.5, info.PriceData.GroupRatioInfo.GroupRatio)
}
