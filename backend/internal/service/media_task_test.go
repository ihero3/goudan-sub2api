package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMediaKindFromModel_Routing(t *testing.T) {
	cases := map[string]MediaKind{
		"wan3.0-video":    MediaKindVideo,
		"wan2.6-t2i":      MediaKindImage,
		"seedance-2.5":    MediaKindVideo,
		"doubao-seedream": MediaKindImage,
		"minimax-h3":      MediaKindVideo,
		"image-01":        MediaKindImage,
		"cosyvoice":       MediaKindAudio,
		"mini-tts":        MediaKindAudio,
	}
	for model, want := range cases {
		got := MediaKindFromModel(model, nil)
		require.Equalf(t, want, got, "model=%q", model)
	}
}

func TestBuildSeedanceImageCreateBody(t *testing.T) {
	req := MediaCreateRequest{UpstreamModel: "doubao-seedream-4-0", Prompt: "a cat", Resolution: "2048x1024"}
	body := buildSeedanceImageCreateBody(req)
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	require.Equal(t, "doubao-seedream-4-0", m["model"])
	require.Equal(t, "a cat", m["prompt"])
	require.Equal(t, "2048x1024", m["size"])
}

func TestBuildWanImageCreateBody(t *testing.T) {
	seed := int64(7)
	req := MediaCreateRequest{UpstreamModel: "wan2.6-t2i", Prompt: "a dog", Resolution: "1280*1280", Seed: &seed}
	body := buildWanImageCreateBody(req)
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	require.Equal(t, "wan2.6-t2i", m["model"])
	input := m["input"].(map[string]any)
	messages := input["messages"].([]any)
	require.Equal(t, "user", messages[0].(map[string]any)["role"])
	params := m["parameters"].(map[string]any)
	require.Equal(t, "1280*1280", params["size"])
	require.Equal(t, float64(7), params["seed"])
}

func TestBuildMiniMaxImageCreateBody_AspectRatio(t *testing.T) {
	req := MediaCreateRequest{UpstreamModel: "image-01", Prompt: "a house", Resolution: "16:9"}
	body := buildMiniMaxImageCreateBody(req)
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	require.Equal(t, "image-01", m["model"])
	require.Equal(t, "16:9", m["aspect_ratio"])
}

func TestBuildSeedanceVideoCreateBody_Content(t *testing.T) {
	req := VideoCreateRequest{UpstreamModel: "doubao-seedance-1-5-pro", Prompt: "sunset"}
	body := buildSeedanceVideoCreateBody(req)
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	require.Equal(t, "doubao-seedance-1-5-pro", m["model"])
	content := m["content"].([]any)
	require.Equal(t, "text", content[0].(map[string]any)["type"])
	require.Equal(t, "sunset", content[0].(map[string]any)["text"])
}
