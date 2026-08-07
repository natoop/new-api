package zzdh

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/fixedmeteredbilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relaydto "github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestModelProfilesExposeAllConfirmedVideoModels(t *testing.T) {
	models := modelList()
	require.Len(t, models, 59)
	protocolCounts := map[protocol]int{}
	for _, name := range models {
		profile, ok := profileForModel(name)
		require.True(t, ok)
		protocolCounts[profile.Protocol]++
	}
	require.Equal(t, 16, protocolCounts[protocolV1])
	require.Equal(t, 43, protocolCounts[protocolV8])
	_, ok := profileForModel("qwen-image-2.0")
	require.False(t, ok)
}

func TestViduQ3UsesDocumentedV8TextToVideoContract(t *testing.T) {
	profile, ok := profileForModel("vidu-q3-pro-720p")
	require.True(t, ok)
	require.Equal(t, protocolV8, profile.Protocol)
	require.True(t, profile.RejectsReference)

	payload, err := buildRequestPayload(&relaycommon.TaskSubmitReq{
		Model:    profile.Name,
		Prompt:   "a corgi runs on a beach",
		Duration: 5,
		Metadata: map[string]interface{}{"resolution": "720P"},
	}, profile)
	require.NoError(t, err)
	require.Equal(t, "720P", payload.Resolution)
	require.NoError(t, validateRequestPayload(payload, profile))

	payload, err = buildRequestPayload(&relaycommon.TaskSubmitReq{
		Model:    profile.Name,
		Prompt:   "a corgi runs on a beach",
		Duration: 5,
		Image:    "https://example.com/reference.jpg",
	}, profile)
	require.NoError(t, err)
	require.Error(t, validateRequestPayload(payload, profile))
}

func TestBuildRequestPayloadEnforcesReferenceVariantRules(t *testing.T) {
	profile, ok := profileForModel("kling-3.0-omni-720p-ref-mute")
	require.True(t, ok)

	payload, err := buildRequestPayload(&relaycommon.TaskSubmitReq{
		Model:    profile.Name,
		Prompt:   "subject turns",
		Duration: 5,
		Images:   []string{"https://example.com/reference.jpg"},
	}, profile)
	require.NoError(t, err)
	require.NoError(t, validateRequestPayload(payload, profile))
	require.Len(t, payload.ReferenceImages, 1)

	noref, ok := profileForModel("kling-3.0-omni-720p-noref-mute")
	require.True(t, ok)
	require.Error(t, validateRequestPayload(payload, noref))
}

func TestBuildRequestPayloadUsesSecondsAndSingleImageCompatibilityFields(t *testing.T) {
	profile, ok := profileForModel("kling-3.0-omni-720p-ref-mute")
	require.True(t, ok)

	payload, err := buildRequestPayload(&relaycommon.TaskSubmitReq{
		Model:   profile.Name,
		Prompt:  "subject turns",
		Seconds: "7",
		Image:   "https://example.com/reference.jpg",
	}, profile)
	require.NoError(t, err)
	require.Equal(t, 7, payload.Duration)
	require.Len(t, payload.ReferenceImages, 1)
	require.Equal(t, "https://example.com/reference.jpg", payload.ReferenceImages[0].URL)
}

func TestMinimaxH3V8PayloadAndBillingRules(t *testing.T) {
	profile, ok := profileForModel("zzdh-Minimax-h3-1080p")
	require.True(t, ok)
	require.Equal(t, protocolV8, profile.Protocol)
	require.Equal(t, "1080p", profile.ResolutionTier)

	req := relaycommon.TaskSubmitReq{
		Model:   profile.Name,
		Prompt:  "cinematic rooftop motion",
		Seconds: "10",
		Metadata: map[string]interface{}{
			"aspect_ratio": "16:9",
			"reference_images": []map[string]interface{}{
				{"url": "data:image/png;base64,AAAA", "role": "reference_image"},
			},
			"extra": map[string]interface{}{"reference_video_audio": false},
		},
	}
	payload, err := buildRequestPayload(&req, profile)
	require.NoError(t, err)
	require.Equal(t, 10, payload.Duration)
	require.Equal(t, 24, *payload.FPS)
	require.Equal(t, "1920x1080", payload.Resolution)
	require.Len(t, payload.ReferenceImages, 1)
	require.Equal(t, "reference_image", payload.ReferenceImages[0].Role)
	require.Equal(t, false, payload.Extra["reference_video_audio"])
	require.NoError(t, validateRequestPayload(payload, profile))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("task_request", req)
	ratio := (&TaskAdaptor{}).EstimateBilling(c, &relaycommon.RelayInfo{OriginModelName: profile.Name})
	require.Equal(t, 10.0, ratio["seconds"])
}

func TestMinimaxH3V8RejectsInvalidDurationFPSAndReferenceMix(t *testing.T) {
	profile, ok := profileForModel("zzdh-Minimax-h3-2k")
	require.True(t, ok)

	payload := &requestPayload{
		Model:      profile.Name,
		Prompt:     "subject",
		Duration:   4,
		FPS:        intPointer(24),
		Resolution: "2560x1440",
	}
	require.Error(t, validateRequestPayload(payload, profile))

	payload.Duration = 10
	payload.FPS = intPointer(30)
	require.Error(t, validateRequestPayload(payload, profile))

	payload.FPS = intPointer(24)
	payload.ReferenceImages = make([]referenceInput, 10)
	for i := range payload.ReferenceImages {
		payload.ReferenceImages[i] = referenceInput{URL: "https://example.com/ref.png", Role: "reference_image"}
	}
	require.Error(t, validateRequestPayload(payload, profile))

	payload.ReferenceImages = []referenceInput{{URL: "https://example.com/last.png", Role: "last_frame"}}
	require.Error(t, validateRequestPayload(payload, profile))

	payload.ReferenceImages = []referenceInput{{URL: "https://example.com/first.png", Role: "first_frame"}}
	payload.ReferenceVideos = []referenceInput{{URL: "https://example.com/ref.mp4"}}
	require.Error(t, validateRequestPayload(payload, profile))
}

func intPointer(value int) *int {
	return &value
}

func TestValidateRequestPayloadRestrictsProfileResolutionAndAspectRatio(t *testing.T) {
	profile, ok := profileForModel("kling-3.0-omni-1080p-noref-mute")
	require.True(t, ok)

	payload := &requestPayload{Prompt: "a subject", Duration: 5, Resolution: "720P"}
	require.Error(t, validateRequestPayload(payload, profile))

	payload.Resolution = "1920x1080"
	payload.AspectRatio = "4:3"
	require.NoError(t, validateRequestPayload(payload, profile))

	payload.AspectRatio = "16:9"
	require.NoError(t, validateRequestPayload(payload, profile))
}

func TestSeedanceV1PayloadKeepsMetadataAndReferenceVideoBilling(t *testing.T) {
	profile, ok := profileForModel("doubao-seedance-2-0-fast-video-480p")
	require.True(t, ok)
	require.Equal(t, protocolV1, profile.Protocol)
	require.True(t, profile.IncludeReferenceVideoDuration)

	req := relaycommon.TaskSubmitReq{
		Model:   profile.Name,
		Prompt:  "continue the reference motion",
		Seconds: "6",
		Metadata: map[string]interface{}{
			"reference_video": "https://example.com/ref.mp4",
			"ratio":           "16:9",
		},
	}
	payload, err := buildRequestPayload(&req, profile)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/ref.mp4", payload.Metadata["reference_video"])
	require.Len(t, payload.ReferenceVideos, 1)
	require.Empty(t, payload.Resolution)
	require.NoError(t, validateRequestPayload(payload, profile))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("task_request", req)
	c.Set(zzdhReferenceKey, float64(2.5))
	adaptor := &TaskAdaptor{}
	ratio := adaptor.EstimateBilling(c, &relaycommon.RelayInfo{OriginModelName: profile.Name})
	require.Equal(t, 8.5, ratio["seconds"])
}

func TestEstimateFixedMeteredBillingUsesReferenceVideoMetric(t *testing.T) {
	profile, ok := profileForModel("doubao-seedance-2-0-fast-video-480p")
	require.True(t, ok)

	req := relaycommon.TaskSubmitReq{Model: profile.Name, Prompt: "continue", Duration: 6}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("task_request", req)
	c.Set(zzdhReferenceKey, float64(2.5))
	c.Set(zzdhFixedUsageModeKey, fixedmeteredbilling.UsageModeDurationWithRef)

	metrics, err := (&TaskAdaptor{}).EstimateFixedMeteredBilling(c, &relaycommon.RelayInfo{})
	require.NoError(t, err)
	require.Equal(t, 6.0, metrics.OutputSeconds)
	require.Equal(t, 2.5, metrics.ReferenceVideoSeconds)
}

func TestSumReferenceVideoSecondsIncludesEveryVideo(t *testing.T) {
	payload := &requestPayload{ReferenceVideos: []referenceInput{
		{URL: "https://example.com/first.mp4"},
		{URL: "https://example.com/second.mp4"},
	}}
	durations := map[string]float64{
		"https://example.com/first.mp4":  2.25,
		"https://example.com/second.mp4": 3.5,
	}

	seconds, err := sumReferenceVideoSeconds(payload, func(url string) (float64, error) {
		return durations[url], nil
	})

	require.NoError(t, err)
	require.Equal(t, 5.75, seconds)
}

func TestBuildRequestURLUsesModelProtocol(t *testing.T) {
	adaptor := &TaskAdaptor{baseURL: "https://zizidonghua.com"}
	v1, ok := profileForModel("doubao-seedance-2-480p")
	require.True(t, ok)
	v8, ok := profileForModel("wan2.7-image")
	require.True(t, ok)
	require.Equal(t, "https://zizidonghua.com/v1/video/generations", adaptorURLForProfile(adaptor, v1))
	require.Equal(t, "https://zizidonghua.com/v8/videos/generations", adaptorURLForProfile(adaptor, v8))
}

func TestFetchTaskUsesStoredModelProtocol(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"queued"}`))
	}))
	defer server.Close()
	adaptor := &TaskAdaptor{}
	_, err := adaptor.FetchTask(server.URL, "key", map[string]any{
		"task_id": "task_v1",
		"model":   "doubao-seedance-2-480p",
	}, "")
	require.NoError(t, err)
	require.Equal(t, "/v1/videos/task_v1", gotPath)
	_, err = adaptor.FetchTask(server.URL, "key", map[string]any{
		"task_id": "task_v8",
		"model":   "wan2.7-image",
	}, "")
	require.NoError(t, err)
	require.Equal(t, "/v8/videos/generations/task_v8", gotPath)
}

func adaptorURLForProfile(adaptor *TaskAdaptor, profile modelProfile) string {
	info := &relaycommon.RelayInfo{OriginModelName: profile.Name}
	url, err := adaptor.BuildRequestURL(info)
	if err != nil {
		return ""
	}
	return url
}

func TestModelListDoesNotIncludeImageOutputModels(t *testing.T) {
	models := strings.Join(modelList(), ",")
	require.NotContains(t, models, "qwen-image-2.0")
	require.NotContains(t, models, "qwen-image-max")
}

func TestParseTaskResultNormalizesV8StatusesAndResultURL(t *testing.T) {
	adaptor := &TaskAdaptor{}
	result, err := adaptor.ParseTaskResult([]byte(`{"id":"task_1","status":"completed","progress":100,"metadata":{"url":"https://cdn.example/video.mp4"}}`))
	require.NoError(t, err)
	require.Equal(t, model.TaskStatusSuccess, result.Status)
	require.Equal(t, "https://cdn.example/video.mp4", result.Url)
	require.Equal(t, "100%", result.Progress)

	result, err = adaptor.ParseTaskResult([]byte(`{"task_id":"task_2","state":"processing","progress":"30%"}`))
	require.NoError(t, err)
	require.Equal(t, model.TaskStatusInProgress, result.Status)
	require.Equal(t, "30%", result.Progress)
}

func TestParseTaskResultRequiresKnownStatus(t *testing.T) {
	adaptor := &TaskAdaptor{}
	_, err := adaptor.ParseTaskResult([]byte(`{"id":"task_1"}`))
	require.Error(t, err)
}

func TestConvertToOpenAIVideoPreservesModelAndInitialQueuedState(t *testing.T) {
	task := &model.Task{
		TaskID:   "task_zzdh_1",
		Status:   model.TaskStatusNotStart,
		Progress: "0%",
		Properties: model.Properties{
			OriginModelName: "doubao-seedance-2-480p",
		},
	}

	body, err := (&TaskAdaptor{}).ConvertToOpenAIVideo(task)
	require.NoError(t, err)
	var response relaydto.OpenAIVideo
	require.NoError(t, common.Unmarshal(body, &response))
	require.Equal(t, task.TaskID, response.ID)
	require.Equal(t, task.TaskID, response.TaskID)
	require.Equal(t, task.Properties.OriginModelName, response.Model)
	require.Equal(t, relaydto.VideoStatusQueued, response.Status)
}
