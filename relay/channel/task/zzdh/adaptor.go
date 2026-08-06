package zzdh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/asynctaskbilling"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relaydto "github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

const (
	zzdhV1SubmitPath  = "/v1/video/generations"
	zzdhV1QueryPath   = "/v1/videos/"
	zzdhV8SubmitPath  = "/v8/videos/generations"
	zzdhV8QueryPath   = "/v8/videos/generations/"
	zzdhReferenceKey  = "zzdh_reference_video_seconds"
	maxReferenceBytes = int64(64 * 1024 * 1024)
)

type referenceInput struct {
	URL  string `json:"url"`
	Role string `json:"role,omitempty"`
}

type mediaURL struct {
	URL string `json:"url,omitempty"`
}

type contentItem struct {
	Type     string    `json:"type,omitempty"`
	ImageURL *mediaURL `json:"image_url,omitempty"`
	VideoURL *mediaURL `json:"video_url,omitempty"`
	AudioURL *mediaURL `json:"audio_url,omitempty"`
	Role     string    `json:"role,omitempty"`
}

type metadataPayload struct {
	Resolution      string           `json:"resolution,omitempty"`
	AspectRatio     string           `json:"aspect_ratio,omitempty"`
	Ratio           string           `json:"ratio,omitempty"`
	FPS             *int             `json:"fps,omitempty"`
	ReferenceVideo  string           `json:"reference_video,omitempty"`
	ReferenceImages []referenceInput `json:"reference_images,omitempty"`
	ReferenceVideos []referenceInput `json:"reference_videos,omitempty"`
	ReferenceAudios []referenceInput `json:"reference_audios,omitempty"`
	Content         []contentItem    `json:"content,omitempty"`
	GenerateAudio   *bool            `json:"generate_audio,omitempty"`
	Seed            *int             `json:"seed,omitempty"`
	NegativePrompt  string           `json:"negative_prompt,omitempty"`
	Extra           map[string]any   `json:"extra,omitempty"`
}

type requestPayload struct {
	Model           string           `json:"model"`
	Prompt          string           `json:"prompt"`
	Duration        int              `json:"duration"`
	FPS             *int             `json:"fps,omitempty"`
	Images          []string         `json:"images,omitempty"`
	Resolution      string           `json:"resolution,omitempty"`
	AspectRatio     string           `json:"aspect_ratio,omitempty"`
	ReferenceImages []referenceInput `json:"reference_images,omitempty"`
	ReferenceVideos []referenceInput `json:"reference_videos,omitempty"`
	ReferenceAudios []referenceInput `json:"reference_audios,omitempty"`
	GenerateAudio   *bool            `json:"generate_audio,omitempty"`
	Seed            *int             `json:"seed,omitempty"`
	NegativePrompt  string           `json:"negative_prompt,omitempty"`
	Extra           map[string]any   `json:"extra,omitempty"`
	Metadata        map[string]any   `json:"metadata,omitempty"`
}

type TaskAdaptor struct {
	taskcommon.BaseBilling
	ChannelType int
	apiKey      string
	baseURL     string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	if info == nil {
		return
	}
	a.ChannelType = info.ChannelType
	a.apiKey = info.ApiKey
	a.baseURL = strings.TrimRight(info.ChannelBaseUrl, "/")
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *taskdto.TaskError {
	if err := relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate); err != nil {
		return err
	}

	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return service.TaskErrorWrapper(err, "get_task_request_failed", http.StatusBadRequest)
	}
	profile, ok := profileForModel(req.Model)
	if !ok {
		return service.TaskErrorWrapperLocal(fmt.Errorf("ZZDH model %q is not enabled", req.Model), "unsupported_model", http.StatusBadRequest)
	}
	asyncTaskBilling := info.AsyncTaskBillingSnapshot != nil || billing_setting.GetBillingMode(req.Model) == billing_setting.BillingModeAsyncTaskExpr
	if asyncTaskBilling {
		if info.AsyncTaskBillingSnapshot == nil {
			config, configured := billing_setting.GetAsyncTaskBilling(req.Model)
			if !configured {
				return service.TaskErrorWrapperLocal(fmt.Errorf("ZZDH model %q requires async task billing terms", req.Model), "async_task_billing_required", http.StatusBadRequest)
			}
			if err := asynctaskbilling.ValidateConfigForRule(profile.AsyncTaskBillingRule(), config); err != nil {
				return service.TaskErrorWrapperLocal(err, "async_task_billing_invalid", http.StatusBadRequest)
			}
		}
	} else if _, configured := ratio_setting.GetModelPrice(req.Model, false); !configured {
		if _, configured = ratio_setting.GetDefaultModelPriceMap()[req.Model]; !configured {
			return service.TaskErrorWrapperLocal(fmt.Errorf("ZZDH model %q requires a configured per-second model price", req.Model), "model_price_required", http.StatusBadRequest)
		}
	}

	payload, err := buildRequestPayload(&req, profile)
	if err != nil {
		return service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
	}
	if err := validateRequestPayload(payload, profile); err != nil {
		return service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
	}
	if profile.BillingRule == billingOutputPlusReferenceSecs {
		referenceURL := firstReferenceVideoURL(payload)
		seconds, err := parseReferenceVideoSeconds(c.Request.Context(), referenceURL)
		if err != nil {
			return service.TaskErrorWrapperLocal(err, "reference_video_duration_failed", http.StatusBadRequest)
		}
		c.Set(zzdhReferenceKey, seconds)
	}
	c.Set("zzdh_profile", profile)
	info.Action = constant.TaskActionGenerate
	return nil
}

func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	if info != nil && (info.AsyncTaskBillingSnapshot != nil || billing_setting.GetBillingMode(info.OriginModelName) == billing_setting.BillingModeAsyncTaskExpr) {
		return nil
	}
	_, metrics, err := a.EstimateAsyncTaskBilling(c, info)
	if err != nil {
		return nil
	}
	seconds := metrics[asynctaskbilling.OutputSeconds]
	if referenceSeconds, ok := metrics[asynctaskbilling.ReferenceVideo]; ok {
		seconds += referenceSeconds
	}
	return map[string]float64{"seconds": seconds}
}

// EstimateAsyncTaskBilling supplies bounded named metrics to the generic
// calculator. Legacy callers keep the single historical seconds multiplier.
func (a *TaskAdaptor) EstimateAsyncTaskBilling(c *gin.Context, _ *relaycommon.RelayInfo) (asynctaskbilling.Rule, map[string]float64, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return asynctaskbilling.Rule{}, nil, err
	}
	profile, ok := profileForModel(req.Model)
	if !ok {
		return asynctaskbilling.Rule{}, nil, fmt.Errorf("ZZDH model %q is not enabled", req.Model)
	}
	payload, err := buildRequestPayload(&req, profile)
	if err != nil || payload.Duration <= 0 {
		if err != nil {
			return asynctaskbilling.Rule{}, nil, err
		}
		return asynctaskbilling.Rule{}, nil, fmt.Errorf("duration is required for async task billing")
	}
	metrics := map[string]float64{asynctaskbilling.OutputSeconds: float64(payload.Duration)}
	if profile.BillingRule == billingOutputPlusReferenceSecs {
		referenceSeconds, ok := c.Get(zzdhReferenceKey)
		if !ok {
			return asynctaskbilling.Rule{}, nil, fmt.Errorf("reference video duration is required for async task billing")
		}
		value, ok := referenceSeconds.(float64)
		if !ok || value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			return asynctaskbilling.Rule{}, nil, fmt.Errorf("reference video duration is outside the supported range")
		}
		metrics[asynctaskbilling.ReferenceVideo] = value
	}
	return profile.AsyncTaskBillingRule(), metrics, nil
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if a.baseURL == "" {
		return "", fmt.Errorf("ZZDH channel base URL is empty")
	}
	profile := profileFromContext(info)
	if profile.Protocol == protocolV1 {
		return a.baseURL + zzdhV1SubmitPath, nil
	}
	return a.baseURL + zzdhV8SubmitPath, nil
}

func (a *TaskAdaptor) BuildRequestHeader(_ *gin.Context, req *http.Request, _ *relaycommon.RelayInfo) error {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil, err
	}
	profile, ok := profileForModel(req.Model)
	if !ok {
		return nil, fmt.Errorf("ZZDH model %q is not enabled", req.Model)
	}
	payload, err := buildRequestPayload(&req, profile)
	if err != nil {
		return nil, err
	}
	if info != nil && info.UpstreamModelName != "" {
		payload.Model = info.UpstreamModelName
	}
	if err := validateRequestPayload(payload, profile); err != nil {
		return nil, err
	}
	body, err := common.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal ZZDH request failed: %w", err)
	}
	return bytes.NewReader(body), nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *taskdto.TaskError) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusBadGateway)
	}
	_ = resp.Body.Close()

	payload, err := decodePayload(body)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "invalid_response", http.StatusBadGateway)
	}
	upstreamID := firstString(payload, "id", "task_id", "taskId", "data.id", "data.task_id", "data.taskId")
	if upstreamID == "" {
		message := responseError(payload)
		if message == "" {
			message = "ZZDH response task id is empty"
		}
		return "", nil, service.TaskErrorWrapper(fmt.Errorf("%s", message), "invalid_response", http.StatusBadGateway)
	}

	response := relaydto.NewOpenAIVideo()
	response.ID = info.PublicTaskID
	response.TaskID = info.PublicTaskID
	response.CreatedAt = time.Now().Unix()
	response.Model = info.OriginModelName
	response.Status = relaydto.VideoStatusQueued
	if status := normalizeStatus(firstString(payload, "status", "state", "task_status", "data.status", "data.state", "data.task_status")); status != "" {
		response.Status = statusToVideoStatus(status)
	}
	c.JSON(http.StatusOK, response)
	return upstreamID, body, nil
}

func (a *TaskAdaptor) FetchTask(baseURL, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok || strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("invalid task_id")
	}
	profile := modelProfile{Protocol: protocolV8}
	if modelName, ok := body["model"].(string); ok {
		if candidate, exists := profileForModel(modelName); exists {
			profile = candidate
		}
	}
	queryPath := zzdhV8QueryPath
	if profile.Protocol == protocolV1 {
		queryPath = zzdhV1QueryPath
	}
	uri := strings.TrimRight(baseURL, "/") + queryPath + url.PathEscape(taskID)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) GetModelList() []string { return modelList() }

func (a *TaskAdaptor) GetChannelName() string { return "zzdh" }

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	payload, err := decodePayload(respBody)
	if err != nil {
		return nil, fmt.Errorf("unmarshal ZZDH task response failed: %w", err)
	}
	status := normalizeStatus(firstString(payload, "status", "state", "task_status", "data.status", "data.state", "data.task_status"))
	if status == "" {
		return nil, fmt.Errorf("ZZDH task status is empty")
	}
	taskInfo := &relaycommon.TaskInfo{
		TaskID:   firstString(payload, "id", "task_id", "taskId", "data.id", "data.task_id", "data.taskId"),
		Progress: progressString(firstValue(payload, "progress", "data.progress")),
		Url: firstString(payload,
			"metadata.url", "url", "video_url", "result_url", "data.metadata.url", "data.url", "data.video_url", "data.result_url",
			"output.url", "output.video_url", "output.video.url", "data.output.url", "data.output.video_url", "data.output.video.url",
			"content.video_url", "data.content.video_url", "result.url", "result.video_url", "data.result.url", "data.result.video_url",
		),
	}
	switch status {
	case "queued":
		taskInfo.Status = model.TaskStatusQueued
	case "in_progress":
		taskInfo.Status = model.TaskStatusInProgress
	case "success":
		taskInfo.Status = model.TaskStatusSuccess
	case "failure":
		taskInfo.Status = model.TaskStatusFailure
		taskInfo.Reason = responseError(payload)
		if taskInfo.Reason == "" {
			taskInfo.Reason = "ZZDH task failed"
		}
	default:
		return nil, fmt.Errorf("unknown ZZDH task status %q", status)
	}
	return taskInfo, nil
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(task *model.Task) ([]byte, error) {
	video := relaydto.NewOpenAIVideo()
	video.ID = task.TaskID
	video.TaskID = task.TaskID
	video.Model = task.Properties.OriginModelName
	video.Status = task.Status.ToVideoStatus()
	video.SetProgressStr(task.Progress)
	video.CreatedAt = task.CreatedAt
	video.CompletedAt = task.UpdatedAt
	if resultURL := task.GetResultURL(); resultURL != "" {
		video.SetMetadata("url", resultURL)
	}
	if task.Status == model.TaskStatusFailure && task.FailReason != "" {
		video.Error = &relaydto.OpenAIVideoError{Message: task.FailReason, Code: "task_failed"}
	}
	return common.Marshal(video)
}

func buildRequestPayload(req *relaycommon.TaskSubmitReq, profile modelProfile) (*requestPayload, error) {
	if req == nil {
		return nil, fmt.Errorf("request is empty")
	}
	payload := &requestPayload{
		Model:    profile.Name,
		Prompt:   strings.TrimSpace(req.Prompt),
		Duration: req.Duration,
	}
	if profile.Protocol == protocolV8 {
		payload.Resolution = profile.DefaultResolution
	}
	if payload.Duration == 0 {
		if rawSeconds := strings.TrimSpace(req.Seconds); rawSeconds != "" {
			parsed, err := strconv.Atoi(rawSeconds)
			if err != nil {
				return nil, fmt.Errorf("seconds must be a valid integer")
			}
			payload.Duration = parsed
		}
		if payload.Duration == 0 {
			payload.Duration = 5
		}
	}

	metadata := metadataPayload{}
	if err := taskcommon.UnmarshalMetadata(req.Metadata, &metadata); err != nil {
		return nil, err
	}
	metadataMap := cloneMetadata(req.Metadata)
	if metadata.Resolution != "" {
		payload.Resolution = metadata.Resolution
	}
	if metadata.AspectRatio != "" {
		payload.AspectRatio = metadata.AspectRatio
	}
	if payload.AspectRatio == "" {
		payload.AspectRatio = metadata.Ratio
	}
	payload.ReferenceImages = append(payload.ReferenceImages, metadata.ReferenceImages...)
	payload.ReferenceVideos = append(payload.ReferenceVideos, metadata.ReferenceVideos...)
	payload.ReferenceAudios = append(payload.ReferenceAudios, metadata.ReferenceAudios...)
	if metadata.ReferenceVideo != "" {
		payload.ReferenceVideos = append(payload.ReferenceVideos, referenceInput{URL: metadata.ReferenceVideo, Role: "reference_video"})
	}
	for _, item := range metadata.Content {
		switch {
		case item.ImageURL != nil && item.ImageURL.URL != "":
			payload.ReferenceImages = append(payload.ReferenceImages, referenceInput{URL: item.ImageURL.URL, Role: item.Role})
		case item.VideoURL != nil && item.VideoURL.URL != "":
			payload.ReferenceVideos = append(payload.ReferenceVideos, referenceInput{URL: item.VideoURL.URL, Role: item.Role})
		case item.AudioURL != nil && item.AudioURL.URL != "":
			payload.ReferenceAudios = append(payload.ReferenceAudios, referenceInput{URL: item.AudioURL.URL, Role: item.Role})
		}
	}
	for index, image := range req.Images {
		if strings.TrimSpace(image) != "" {
			image = strings.TrimSpace(image)
			payload.Images = append(payload.Images, image)
			role := "first_frame"
			if index == 1 {
				role = "last_frame"
			}
			payload.ReferenceImages = append(payload.ReferenceImages, referenceInput{URL: image, Role: role})
		}
	}
	if len(req.Images) == 0 && strings.TrimSpace(req.Image) != "" {
		image := strings.TrimSpace(req.Image)
		payload.Images = append(payload.Images, image)
		payload.ReferenceImages = append(payload.ReferenceImages, referenceInput{URL: image, Role: "first_frame"})
	}
	if strings.TrimSpace(req.InputReference) != "" && len(payload.ReferenceImages) == 0 {
		image := strings.TrimSpace(req.InputReference)
		payload.Images = append(payload.Images, image)
		payload.ReferenceImages = append(payload.ReferenceImages, referenceInput{URL: image, Role: "first_frame"})
	}
	payload.GenerateAudio = metadata.GenerateAudio
	payload.Seed = metadata.Seed
	payload.NegativePrompt = metadata.NegativePrompt
	if profile.Family == "minimax_h3_v8" {
		payload.FPS = metadata.FPS
		payload.Extra = metadata.Extra
		if payload.FPS == nil {
			fps := profile.FixedFPS
			payload.FPS = &fps
		}
	}
	if payload.GenerateAudio == nil {
		if profile.RequiresAudio {
			value := true
			payload.GenerateAudio = &value
		} else if profile.RejectsAudio {
			value := false
			payload.GenerateAudio = &value
		}
	}
	if profile.Protocol == protocolV1 {
		payload.Metadata = metadataMap
	}
	return payload, nil
}

func validateRequestPayload(payload *requestPayload, profile modelProfile) error {
	if payload == nil {
		return fmt.Errorf("request payload is empty")
	}
	if strings.TrimSpace(payload.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	minDuration := 1
	if profile.MinDuration > 0 {
		minDuration = profile.MinDuration
	}
	maxDuration := relaycommon.MaxTaskDurationSeconds
	if profile.MaxDuration > 0 && profile.MaxDuration < maxDuration {
		maxDuration = profile.MaxDuration
	}
	if payload.Duration < minDuration || payload.Duration > maxDuration {
		return fmt.Errorf("duration must be between %d and %d seconds", minDuration, maxDuration)
	}
	if payload.Resolution != "" && !validResolution(payload.Resolution, profile.ResolutionTier) {
		return fmt.Errorf("model %s requires %s resolution", profile.Name, profile.ResolutionTier)
	}
	if payload.AspectRatio != "" && !validAspectRatio(payload.AspectRatio) {
		return fmt.Errorf("aspect_ratio must be one of 16:9, 9:16, 1:1, 4:3, 3:4, or 21:9")
	}
	if profile.FixedFPS > 0 {
		if payload.FPS == nil || *payload.FPS != profile.FixedFPS {
			return fmt.Errorf("model %s requires %d fps", profile.Name, profile.FixedFPS)
		}
	}
	hasImage := len(payload.ReferenceImages) > 0
	hasVideo := len(payload.ReferenceVideos) > 0
	hasReference := hasImage || hasVideo || len(payload.ReferenceAudios) > 0
	if profile.RequiresReference && !hasReference {
		return fmt.Errorf("model %s requires reference input", profile.Name)
	}
	if profile.RequiresImage && !hasImage {
		return fmt.Errorf("model %s requires a reference image", profile.Name)
	}
	if profile.RequiresVideo && !hasVideo {
		return fmt.Errorf("model %s requires a reference video", profile.Name)
	}
	if profile.RejectsReference && hasReference {
		return fmt.Errorf("model %s does not accept reference input", profile.Name)
	}
	if profile.RejectsVideo && hasVideo {
		return fmt.Errorf("model %s does not accept reference video", profile.Name)
	}
	inputs := append(append(append([]referenceInput{}, payload.ReferenceImages...), payload.ReferenceVideos...), payload.ReferenceAudios...)
	for _, input := range inputs {
		if err := validateReferenceURL(input.URL, profile.AllowDataReference); err != nil {
			return err
		}
	}
	if profile.Family == "minimax_h3_v8" {
		if err := validateMinimaxH3References(payload, profile); err != nil {
			return err
		}
	}
	if payload.GenerateAudio != nil {
		if profile.RequiresAudio && !*payload.GenerateAudio {
			return fmt.Errorf("model %s requires audio", profile.Name)
		}
		if profile.RejectsAudio && *payload.GenerateAudio {
			return fmt.Errorf("model %s does not accept audio", profile.Name)
		}
	}
	if payload.Seed != nil && *payload.Seed < 0 {
		return fmt.Errorf("seed must be non-negative")
	}
	return nil
}

func validResolution(raw, tier string) bool {
	resolution := strings.ToLower(strings.TrimSpace(raw))
	if resolution == "" {
		return true
	}
	if tier == "" {
		return resolution == "480p" || resolution == "720p" || resolution == "1080p" || resolution == "4k" ||
			resolution == "854x480" || resolution == "480x854" || resolution == "1280x720" || resolution == "720x1280" ||
			resolution == "1920x1080" || resolution == "1080x1920" || resolution == "3840x2160" || resolution == "2160x3840" ||
			resolution == "720x720" || resolution == "1080x1080"
	}
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case "480p":
		return resolution == "480p" || resolution == "854x480" || resolution == "480x854"
	case "720p":
		return resolution == "720p" || resolution == "1280x720" || resolution == "720x1280" || resolution == "720x720"
	case "1080p":
		return resolution == "1080p" || resolution == "1920x1080" || resolution == "1080x1920" || resolution == "1080x1080"
	case "4k":
		return resolution == "4k" || resolution == "3840x2160" || resolution == "2160x3840"
	case "2k":
		return resolution == "2k" || resolution == "2560x1440" || resolution == "1440x2560"
	default:
		return false
	}
}

func validAspectRatio(raw string) bool {
	switch strings.TrimSpace(raw) {
	case "16:9", "9:16", "1:1", "4:3", "3:4", "21:9":
		return true
	default:
		return false
	}
}

func validateReferenceURL(raw string, allowData bool) error {
	raw = strings.TrimSpace(raw)
	if allowData && strings.HasPrefix(strings.ToLower(raw), "data:") {
		if !strings.Contains(raw, ",") {
			return fmt.Errorf("data reference input must contain a payload")
		}
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("reference input must be an absolute http or https URL")
	}
	return nil
}

func validateMinimaxH3References(payload *requestPayload, profile modelProfile) error {
	if len(payload.ReferenceImages) > profile.MaxReferenceImages {
		return fmt.Errorf("model %s accepts at most %d reference images", profile.Name, profile.MaxReferenceImages)
	}
	if len(payload.ReferenceVideos) > profile.MaxReferenceVideos {
		return fmt.Errorf("model %s accepts at most %d reference videos", profile.Name, profile.MaxReferenceVideos)
	}
	if len(payload.ReferenceAudios) > profile.MaxReferenceAudios {
		return fmt.Errorf("model %s accepts at most %d reference audios", profile.Name, profile.MaxReferenceAudios)
	}

	frameCount := 0
	referenceImageCount := 0
	blankRoleCount := 0
	firstFrame := false
	lastFrame := false
	for _, input := range payload.ReferenceImages {
		switch strings.ToLower(strings.TrimSpace(input.Role)) {
		case "":
			blankRoleCount++
		case "first_frame":
			frameCount++
			firstFrame = true
		case "last_frame":
			frameCount++
			lastFrame = true
		case "reference_image":
			referenceImageCount++
		default:
			return fmt.Errorf("model %s does not support image role %q", profile.Name, input.Role)
		}
	}
	if frameCount > 0 {
		if blankRoleCount > 0 || referenceImageCount > 0 || len(payload.ReferenceImages) > 2 {
			return fmt.Errorf("frame references cannot be mixed with other image roles")
		}
		if lastFrame && !firstFrame {
			return fmt.Errorf("last_frame requires first_frame")
		}
		if len(payload.ReferenceVideos) > 0 || len(payload.ReferenceAudios) > 0 {
			return fmt.Errorf("frame references cannot be mixed with reference video or audio")
		}
	}
	if referenceImageCount > 0 && blankRoleCount > 0 {
		return fmt.Errorf("reference_image role must be set on every reference image")
	}
	for _, input := range payload.ReferenceVideos {
		role := strings.ToLower(strings.TrimSpace(input.Role))
		if role != "" && role != "reference_video" {
			return fmt.Errorf("model %s does not support video role %q", profile.Name, input.Role)
		}
	}
	for _, input := range payload.ReferenceAudios {
		role := strings.ToLower(strings.TrimSpace(input.Role))
		if role != "" && role != "reference_audio" {
			return fmt.Errorf("model %s does not support audio role %q", profile.Name, input.Role)
		}
	}
	return nil
}

func firstReferenceVideoURL(payload *requestPayload) string {
	if payload == nil {
		return ""
	}
	for _, input := range payload.ReferenceVideos {
		if strings.TrimSpace(input.URL) != "" {
			return strings.TrimSpace(input.URL)
		}
	}
	return ""
}

func parseReferenceVideoSeconds(ctx context.Context, rawURL string) (float64, error) {
	if err := validateReferenceURL(rawURL, false); err != nil {
		return 0, err
	}
	resp, err := service.DoDownloadRequest(rawURL, "zzdh_reference_video_duration")
	if err != nil {
		return 0, fmt.Errorf("download reference video failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return 0, fmt.Errorf("reference video returned status %d", resp.StatusCode)
	}
	limit := maxReferenceBytes
	if constant.MaxFileDownloadMB > 0 {
		limit = int64(constant.MaxFileDownloadMB) * 1024 * 1024
	}
	if resp.ContentLength > limit {
		return 0, fmt.Errorf("reference video exceeds the %d MB download limit", limit/(1024*1024))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return 0, fmt.Errorf("read reference video failed: %w", err)
	}
	if int64(len(body)) > limit {
		return 0, fmt.Errorf("reference video exceeds the %d MB download limit", limit/(1024*1024))
	}
	ext := strings.ToLower(filepath.Ext(strings.Split(rawURL, "?")[0]))
	if ext == "" {
		ext = ".mp4"
	}
	if !contains([]string{".mp4", ".m4a", ".webm"}, ext) {
		contentType := strings.ToLower(resp.Header.Get("Content-Type"))
		if strings.Contains(contentType, "webm") {
			ext = ".webm"
		} else {
			ext = ".mp4"
		}
	}
	seconds, err := common.GetAudioDuration(ctx, bytes.NewReader(body), ext)
	if err != nil {
		return 0, fmt.Errorf("parse reference video duration failed: %w", err)
	}
	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds > relaycommon.MaxTaskDurationSeconds {
		return 0, fmt.Errorf("reference video duration is outside the supported range")
	}
	return seconds, nil
}

func cloneMetadata(metadata map[string]interface{}) map[string]any {
	if len(metadata) == 0 {
		return nil
	}
	body, err := common.Marshal(metadata)
	if err != nil {
		return nil
	}
	var clone map[string]any
	if err := common.Unmarshal(body, &clone); err != nil {
		return nil
	}
	delete(clone, "model")
	return clone
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func profileFromContext(info *relaycommon.RelayInfo) modelProfile {
	if info != nil && info.OriginModelName != "" {
		if profile, ok := profileForModel(info.OriginModelName); ok {
			return profile
		}
	}
	return modelProfile{Protocol: protocolV8}
}

func decodePayload(body []byte) (map[string]any, error) {
	var payload map[string]any
	if err := common.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func firstValue(payload map[string]any, paths ...string) any {
	for _, path := range paths {
		if value, ok := valueAtPath(payload, path); ok && value != nil {
			return value
		}
	}
	return nil
}

func firstString(payload map[string]any, paths ...string) string {
	value := firstValue(payload, paths...)
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	default:
		return ""
	}
}

func valueAtPath(root any, path string) (any, bool) {
	current := root
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func responseError(payload map[string]any) string {
	return firstString(payload,
		"error.message", "message", "error", "data.error.message", "data.message", "data.error",
	)
}

func normalizeStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "queued", "pending", "created", "queueing", "submitted":
		return "queued"
	case "processing", "in_progress", "running", "in-progress":
		return "in_progress"
	case "completed", "complete", "succeeded", "success", "succeed":
		return "success"
	case "failed", "failure", "error", "cancelled", "canceled":
		return "failure"
	default:
		return ""
	}
}

func statusToVideoStatus(status string) string {
	switch normalizeStatus(status) {
	case "queued":
		return relaydto.VideoStatusQueued
	case "in_progress":
		return relaydto.VideoStatusInProgress
	case "success":
		return relaydto.VideoStatusCompleted
	case "failure":
		return relaydto.VideoStatusFailed
	default:
		return relaydto.VideoStatusUnknown
	}
}

func progressString(value any) string {
	switch typed := value.(type) {
	case string:
		if strings.HasSuffix(typed, "%") {
			return typed
		}
		return typed + "%"
	case float64:
		return strconv.Itoa(int(typed)) + "%"
	case int:
		return strconv.Itoa(typed) + "%"
	default:
		return ""
	}
}
