package video_outreach

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectTemplate_Tier1(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	attribs := map[string]string{"lead_tier": "tier_1"}

	sel := SelectTemplate(cfg, attribs)

	assert.Equal(t, 10, sel.TemplateID)
	assert.Equal(t, "video", sel.Type)
	assert.Equal(t, "tier_1", sel.Tier)
}

func TestSelectTemplate_Tier2(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	attribs := map[string]string{"lead_tier": "tier_2"}

	sel := SelectTemplate(cfg, attribs)

	assert.Equal(t, 20, sel.TemplateID)
	assert.Equal(t, "text", sel.Type)
	assert.Equal(t, "tier_2", sel.Tier)
}

func TestSelectTemplate_NoTier(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	attribs := map[string]string{}

	sel := SelectTemplate(cfg, attribs)

	assert.Equal(t, 0, sel.TemplateID)
	assert.Equal(t, "skip", sel.Type)
	assert.Empty(t, sel.Tier)
}

func TestSelectTemplate_UnknownTier(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	attribs := map[string]string{"lead_tier": "tier_3"}

	sel := SelectTemplate(cfg, attribs)

	assert.Equal(t, "skip", sel.Type)
}

func TestSelectTemplate_NilAttribs(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}

	sel := SelectTemplate(cfg, map[string]string{})

	assert.Equal(t, "skip", sel.Type)
}

func TestHasVideoAssets_AllPresent(t *testing.T) {
	attribs := map[string]string{
		"video_url":        "https://cdn.example.com/video.mp4",
		"thumbnail_url":    "https://cdn.example.com/thumb.png",
		"landing_page_url": "https://example.com/watch?v=123",
	}
	assert.True(t, HasVideoAssets(attribs))
}

func TestHasVideoAssets_MissingOne(t *testing.T) {
	tests := []struct {
		name    string
		attribs map[string]string
	}{
		{"missing video", map[string]string{"thumbnail_url": "x", "landing_page_url": "x"}},
		{"missing thumb", map[string]string{"video_url": "x", "landing_page_url": "x"}},
		{"missing landing", map[string]string{"video_url": "x", "thumbnail_url": "x"}},
		{"all empty", map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, HasVideoAssets(tt.attribs))
		})
	}
}

func TestIsVideoEligible_Tier1WithAssets(t *testing.T) {
	attribs := map[string]string{
		"lead_tier":        "tier_1",
		"video_url":        "https://cdn.example.com/video.mp4",
		"thumbnail_url":    "https://cdn.example.com/thumb.png",
		"landing_page_url": "https://example.com/watch",
	}
	assert.True(t, IsVideoEligible(attribs))
}

func TestIsVideoEligible_Tier1NoAssets(t *testing.T) {
	attribs := map[string]string{"lead_tier": "tier_1"}
	assert.False(t, IsVideoEligible(attribs))
}

func TestIsVideoEligible_Tier2WithAssets(t *testing.T) {
	attribs := map[string]string{
		"lead_tier":        "tier_2",
		"video_url":        "https://cdn.example.com/video.mp4",
		"thumbnail_url":    "https://cdn.example.com/thumb.png",
		"landing_page_url": "https://example.com/watch",
	}
	assert.False(t, IsVideoEligible(attribs))
}
