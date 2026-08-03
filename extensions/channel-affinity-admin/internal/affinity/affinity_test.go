package affinity

import (
	"testing"

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
