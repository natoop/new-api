package affinity

import (
	"context"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestKeyMatchesNewAPIChannelAffinityProtocol(t *testing.T) {
	key := Key(Binding{UserID: 42, Group: "vip", Model: "gpt-4.1", ChannelID: 9})
	require.Equal(t, "new-api:channel_affinity:v1:external-admin-user-channel-v1:gpt-4.1:vip:42", key)
}

func TestValidateRejectsAmbiguousKeyPartsAndAutoGroupByDefault(t *testing.T) {
	require.Error(t, Validate(Binding{UserID: 1, ChannelID: 2, Group: "auto", Model: "gpt-4.1"}, true))
	require.Error(t, Validate(Binding{UserID: 1, ChannelID: 2, Group: "vip:prod", Model: "gpt-4.1"}, true))
	require.Error(t, Validate(Binding{UserID: 1, ChannelID: 2, Group: "vip", Model: ""}, true))
	require.Error(t, Validate(Binding{UserID: 1, Group: "vip", Model: "gpt-4.1"}, true))
	require.NoError(t, Validate(Binding{UserID: 1, ChannelID: 2, Group: "VIP 会员", Model: "gpt-4.1"}, true))
}

func TestDeleteByChannelRemovesOnlyMatchingAffinityValuesAndIsIdempotent(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	previousClient := common.RDB
	previousEnabled := common.RedisEnabled
	common.RDB = client
	common.RedisEnabled = true
	t.Cleanup(func() {
		common.RDB = previousClient
		common.RedisEnabled = previousEnabled
		require.NoError(t, client.Close())
	})

	ctx := context.Background()
	matchingExternal := Namespace + ":external-admin-user-channel-v1:gpt-4.1:vip:1"
	matchingOtherRule := Namespace + ":prompt-cache:gpt-4.1:vip:key"
	otherChannel := Namespace + ":external-admin-user-channel-v1:gpt-4.1:vip:2"
	require.NoError(t, client.Set(ctx, matchingExternal, "9", 0).Err())
	require.NoError(t, client.Set(ctx, matchingOtherRule, "9", 0).Err())
	require.NoError(t, client.Set(ctx, otherChannel, "10", 0).Err())
	require.NoError(t, client.Set(ctx, "unrelated:key", "9", 0).Err())

	result, err := DeleteByChannel(ctx, 9, ChannelClearAuditEvent{ChannelID: 9, ActorID: 1, RequestID: "request-1"})
	require.NoError(t, err)
	require.EqualValues(t, 3, result.Scanned)
	require.EqualValues(t, 2, result.Deleted)
	require.Equal(t, int64(0), client.Exists(ctx, matchingExternal, matchingOtherRule).Val())
	require.Equal(t, "10", client.Get(ctx, otherChannel).Val())
	require.Equal(t, "9", client.Get(ctx, "unrelated:key").Val())

	retry, err := DeleteByChannel(ctx, 9, ChannelClearAuditEvent{ChannelID: 9, ActorID: 1, RequestID: "request-2"})
	require.NoError(t, err)
	require.Zero(t, retry.Deleted)
	require.Equal(t, "10", client.Get(ctx, otherChannel).Val())
	require.Len(t, client.XRange(ctx, auditStream, "-", "+").Val(), 2)
}
