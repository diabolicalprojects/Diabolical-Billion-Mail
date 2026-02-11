package video_outreach

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ===== SelectTemplate =====

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
	assert.Empty(t, sel.Tier)
	assert.Zero(t, sel.TemplateID)
}

func TestSelectTemplate_EmptyTierValue(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	attribs := map[string]string{"lead_tier": ""}

	sel := SelectTemplate(cfg, attribs)
	assert.Equal(t, "skip", sel.Type)
}

func TestSelectTemplate_NilAttribs(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}

	sel := SelectTemplate(cfg, map[string]string{})

	assert.Equal(t, "skip", sel.Type)
}

func TestSelectTemplate_ZeroTemplateIDs(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 0, TextTemplateID: 0}
	attribs := map[string]string{"lead_tier": "tier_1"}

	sel := SelectTemplate(cfg, attribs)

	assert.Equal(t, 0, sel.TemplateID)
	assert.Equal(t, "video", sel.Type)
	assert.Equal(t, "tier_1", sel.Tier)
}

func TestSelectTemplate_CaseSensitivity(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}

	// tier values are case-sensitive
	tests := []struct {
		tier string
		want string
	}{
		{"tier_1", "video"},
		{"Tier_1", "skip"},
		{"TIER_1", "skip"},
		{"tier_2", "text"},
		{"Tier_2", "skip"},
	}
	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			sel := SelectTemplate(cfg, map[string]string{"lead_tier": tt.tier})
			assert.Equal(t, tt.want, sel.Type)
		})
	}
}

func TestSelectTemplate_ExtraAttribsIgnored(t *testing.T) {
	cfg := VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	attribs := map[string]string{
		"lead_tier":   "tier_1",
		"lead_score":  "85",
		"video_url":   "https://cdn.example.com/video.mp4",
		"random_key":  "random_value",
	}

	sel := SelectTemplate(cfg, attribs)
	assert.Equal(t, "video", sel.Type)
	assert.Equal(t, 10, sel.TemplateID)
}

// ===== HasVideoAssets =====

func TestHasVideoAssets_AllPresent(t *testing.T) {
	attribs := map[string]string{
		"video_url":        "https://cdn.example.com/video.mp4",
		"thumbnail_url":    "https://cdn.example.com/thumb.png",
		"landing_page_url": "https://example.com/watch?v=123",
	}
	assert.True(t, HasVideoAssets(attribs))
}

func TestHasVideoAssets_MissingEach(t *testing.T) {
	tests := []struct {
		name    string
		attribs map[string]string
	}{
		{"missing video", map[string]string{"thumbnail_url": "x", "landing_page_url": "x"}},
		{"missing thumb", map[string]string{"video_url": "x", "landing_page_url": "x"}},
		{"missing landing", map[string]string{"video_url": "x", "thumbnail_url": "x"}},
		{"all empty strings", map[string]string{"video_url": "", "thumbnail_url": "", "landing_page_url": ""}},
		{"all missing", map[string]string{}},
		{"only video", map[string]string{"video_url": "x"}},
		{"only thumb", map[string]string{"thumbnail_url": "x"}},
		{"only landing", map[string]string{"landing_page_url": "x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, HasVideoAssets(tt.attribs))
		})
	}
}

func TestHasVideoAssets_EmptyStringValues(t *testing.T) {
	attribs := map[string]string{
		"video_url":        "",
		"thumbnail_url":    "https://cdn.example.com/thumb.png",
		"landing_page_url": "https://example.com/watch",
	}
	assert.False(t, HasVideoAssets(attribs))
}

func TestHasVideoAssets_ExtraKeysIgnored(t *testing.T) {
	attribs := map[string]string{
		"video_url":        "https://cdn.example.com/video.mp4",
		"thumbnail_url":    "https://cdn.example.com/thumb.png",
		"landing_page_url": "https://example.com/watch",
		"lead_tier":        "tier_1",
		"lead_score":       "90",
	}
	assert.True(t, HasVideoAssets(attribs))
}

// ===== IsVideoEligible =====

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

func TestIsVideoEligible_Tier1PartialAssets(t *testing.T) {
	attribs := map[string]string{
		"lead_tier":     "tier_1",
		"video_url":     "https://cdn.example.com/video.mp4",
		"thumbnail_url": "https://cdn.example.com/thumb.png",
		// missing landing_page_url
	}
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

func TestIsVideoEligible_NoTier(t *testing.T) {
	attribs := map[string]string{
		"video_url":        "https://cdn.example.com/video.mp4",
		"thumbnail_url":    "https://cdn.example.com/thumb.png",
		"landing_page_url": "https://example.com/watch",
	}
	assert.False(t, IsVideoEligible(attribs))
}

func TestIsVideoEligible_EmptyAttribs(t *testing.T) {
	assert.False(t, IsVideoEligible(map[string]string{}))
}

func TestIsVideoEligible_EmptyTierWithAssets(t *testing.T) {
	attribs := map[string]string{
		"lead_tier":        "",
		"video_url":        "x",
		"thumbnail_url":    "x",
		"landing_page_url": "x",
	}
	assert.False(t, IsVideoEligible(attribs))
}

// ===== Constants =====

func TestAttrConstants(t *testing.T) {
	assert.Equal(t, "video_url", AttrVideoURL)
	assert.Equal(t, "thumbnail_url", AttrThumbnailURL)
	assert.Equal(t, "landing_page_url", AttrLandingPageURL)
}

// ===== TemplateSelection struct =====

func TestTemplateSelection_ZeroValue(t *testing.T) {
	var sel TemplateSelection
	assert.Zero(t, sel.TemplateID)
	assert.Empty(t, sel.Type)
	assert.Empty(t, sel.Tier)
}

// ===== VideoOutreachConfig struct =====

func TestVideoOutreachConfig_ZeroValue(t *testing.T) {
	var cfg VideoOutreachConfig
	assert.Zero(t, cfg.VideoTemplateID)
	assert.Zero(t, cfg.TextTemplateID)
}
