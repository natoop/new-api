package relay

import (
	"testing"

	"github.com/QuantumNous/new-api/pkg/fixedmeteredbilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type frozenFixedMeteredBillingEstimator struct {
	calls int
}

func (e *frozenFixedMeteredBillingEstimator) EstimateFixedMeteredBilling(_ *gin.Context, _ *relaycommon.RelayInfo) (fixedmeteredbilling.Metrics, error) {
	e.calls++
	return fixedmeteredbilling.Metrics{}, nil
}

func TestPrepareFixedMeteredBillingReusesFrozenSnapshot(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	estimator := &frozenFixedMeteredBillingEstimator{}
	info := &relaycommon.RelayInfo{
		OriginModelName: "fixed-retry-model",
		FixedMeteredBillingSnapshot: &fixedmeteredbilling.Snapshot{
			BillingMode:   fixedmeteredbilling.BillingMode,
			GroupRatio:    1.5,
			ReservedQuota: 375000,
		},
	}

	used, taskErr := prepareFixedMeteredBilling(ctx, info, estimator)

	require.Nil(t, taskErr)
	assert.True(t, used)
	assert.Zero(t, estimator.calls)
	assert.Equal(t, 375000, info.PriceData.Quota)
	assert.Equal(t, 1.5, info.PriceData.GroupRatioInfo.GroupRatio)
}
