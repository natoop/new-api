package relay

import (
	"fmt"
	"net/http"

	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/pkg/fixedmeteredbilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

// fixedMeteredBillingEstimator is implemented by task adaptors that can
// provide verified duration metrics. Per-request pricing needs no adaptor data.
type fixedMeteredBillingEstimator interface {
	EstimateFixedMeteredBilling(c *gin.Context, info *relaycommon.RelayInfo) (fixedmeteredbilling.Metrics, error)
}

// prepareFixedMeteredBilling selects system-level fixed-metered pricing before
// legacy task pricing. A snapshot is frozen on the first attempt and reused by
// retries so config changes cannot alter an accepted task's reserve.
func prepareFixedMeteredBilling(c *gin.Context, info *relaycommon.RelayInfo, adaptor any) (bool, *taskdto.TaskError) {
	if info.FixedMeteredBillingSnapshot == nil && billing_setting.GetBillingMode(info.OriginModelName) != billing_setting.BillingModeFixedMetered {
		return false, nil
	}

	snapshot := info.FixedMeteredBillingSnapshot
	if snapshot == nil {
		config, ok := billing_setting.GetFixedMeteredBilling(info.OriginModelName)
		if !ok {
			return true, service.TaskErrorWrapperLocal(fmt.Errorf("model %q requires fixed metered billing configuration", info.OriginModelName), "fixed_metered_billing_required", http.StatusBadRequest)
		}
		metrics := fixedmeteredbilling.Metrics{}
		if config.UsageMode != fixedmeteredbilling.UsageModePerRequest {
			estimator, ok := adaptor.(fixedMeteredBillingEstimator)
			if !ok {
				return true, service.TaskErrorWrapperLocal(fmt.Errorf("model %q cannot provide fixed metered duration metrics", info.OriginModelName), "fixed_metered_metrics_unsupported", http.StatusBadRequest)
			}
			var err error
			metrics, err = estimator.EstimateFixedMeteredBilling(c, info)
			if err != nil {
				return true, service.TaskErrorWrapperLocal(err, "fixed_metered_metrics_invalid", http.StatusBadRequest)
			}
		}
		if _, err := helper.FixedMeteredPriceHelper(c, info, metrics); err != nil {
			return true, service.TaskErrorWrapperLocal(err, "fixed_metered_billing_invalid", http.StatusBadRequest)
		}
		snapshot = info.FixedMeteredBillingSnapshot
	}
	if snapshot == nil {
		return true, service.TaskErrorWrapperLocal(fmt.Errorf("model %q did not produce a fixed metered billing snapshot", info.OriginModelName), "fixed_metered_snapshot_missing", http.StatusInternalServerError)
	}

	info.PriceData = helper.FixedMeteredPriceDataFromSnapshot(snapshot)
	if !operation_setting.GetQuotaSetting().EnableFreeModelPreConsume && snapshot.ReservedQuota == 0 {
		info.PriceData.FreeModel = true
	}
	return true, nil
}
