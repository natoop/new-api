package relay

import (
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/pkg/asynctaskbilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	hosttypes "github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// asyncTaskBillingEstimator is deliberately local to task relay. Existing
// adaptors remain unchanged; only providers that opt in expose named metrics.
type asyncTaskBillingEstimator interface {
	EstimateAsyncTaskBilling(c *gin.Context, info *relaycommon.RelayInfo) (asynctaskbilling.Rule, map[string]float64, error)
}

// prepareAsyncTaskBilling selects the independent task calculator before the
// legacy ModelPrice path. The first attempt freezes the rule, terms, metrics,
// group ratio, and quota; every retry reuses that snapshot and reservation.
func prepareAsyncTaskBilling(c *gin.Context, info *relaycommon.RelayInfo, adaptor any) (bool, *taskdto.TaskError) {
	if info.AsyncTaskBillingSnapshot == nil && billing_setting.GetBillingMode(info.OriginModelName) != billing_setting.BillingModeAsyncTaskExpr {
		return false, nil
	}

	snapshot := info.AsyncTaskBillingSnapshot
	if snapshot == nil {
		estimator, ok := adaptor.(asyncTaskBillingEstimator)
		if !ok {
			return true, service.TaskErrorWrapperLocal(fmt.Errorf("model %q does not support async task billing", info.OriginModelName), "async_task_billing_unsupported", http.StatusBadRequest)
		}
		config, ok := billing_setting.GetAsyncTaskBilling(info.OriginModelName)
		if !ok {
			return true, service.TaskErrorWrapperLocal(fmt.Errorf("model %q requires async task billing terms", info.OriginModelName), "async_task_billing_required", http.StatusBadRequest)
		}
		rule, metrics, err := estimator.EstimateAsyncTaskBilling(c, info)
		if err != nil {
			return true, service.TaskErrorWrapperLocal(err, "async_task_billing_metrics_invalid", http.StatusBadRequest)
		}
		groupRatioInfo := helper.HandleGroupRatio(c, info)
		result, err := asynctaskbilling.Calculate(rule, config, metrics, groupRatioInfo.GroupRatio, common.QuotaPerUnit)
		if result != nil && result.Clamp != nil {
			noteTaskQuotaClamp(info, result.Clamp)
		}
		if err != nil {
			return true, service.TaskErrorWrapperLocal(err, "async_task_billing_invalid", http.StatusBadRequest)
		}
		snapshot = &result.Snapshot
		info.AsyncTaskBillingSnapshot = snapshot
	}

	freeModel := !operation_setting.GetQuotaSetting().EnableFreeModelPreConsume && snapshot.GroupRatio == 0
	info.PriceData = hosttypes.PriceData{
		FreeModel:         freeModel,
		UsePrice:          true,
		Quota:             snapshot.ReservedQuota,
		QuotaToPreConsume: snapshot.ReservedQuota,
		GroupRatioInfo: hosttypes.GroupRatioInfo{
			GroupRatio: snapshot.GroupRatio,
		},
	}
	return true, nil
}
