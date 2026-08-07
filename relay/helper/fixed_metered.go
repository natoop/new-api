package helper

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/fixedmeteredbilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	hosttypes "github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// FixedMeteredPriceHelper resolves the complete system-level fixed-metered
// configuration for one public model. It deliberately never falls back to
// ModelPrice or ratio pricing when the mode was selected.
func FixedMeteredPriceHelper(c *gin.Context, info *relaycommon.RelayInfo, metrics fixedmeteredbilling.Metrics) (hosttypes.PriceData, error) {
	config, ok := billing_setting.GetFixedMeteredBilling(info.OriginModelName)
	if !ok {
		return hosttypes.PriceData{}, fmt.Errorf("model %s is configured as fixed_metered but has no fixed metered billing configuration", info.OriginModelName)
	}
	groupRatioInfo := HandleGroupRatio(c, info)
	result, err := fixedmeteredbilling.Calculate(config, metrics, groupRatioInfo.GroupRatio, common.QuotaPerUnit)
	if result != nil && result.Clamp != nil {
		info.QuotaClamp = result.Clamp
	}
	if err != nil {
		return hosttypes.PriceData{}, err
	}

	freeModel := false
	if !operation_setting.GetQuotaSetting().EnableFreeModelPreConsume && (groupRatioInfo.GroupRatio == 0 || config.UnitPrice == 0) {
		freeModel = true
	}
	info.FixedMeteredBillingSnapshot = &result.Snapshot
	priceData := hosttypes.PriceData{
		FreeModel:         freeModel,
		ModelPrice:        config.UnitPrice,
		UsePrice:          true,
		Quota:             result.Quota,
		QuotaToPreConsume: result.Quota,
		GroupRatioInfo:    groupRatioInfo,
	}
	info.PriceData = priceData
	return priceData, nil
}

func FixedMeteredPriceDataFromSnapshot(snapshot *fixedmeteredbilling.Snapshot) hosttypes.PriceData {
	if snapshot == nil {
		return hosttypes.PriceData{}
	}
	return hosttypes.PriceData{
		UsePrice:          true,
		Quota:             snapshot.ReservedQuota,
		QuotaToPreConsume: snapshot.ReservedQuota,
		GroupRatioInfo: hosttypes.GroupRatioInfo{
			GroupRatio: snapshot.GroupRatio,
		},
	}
}
