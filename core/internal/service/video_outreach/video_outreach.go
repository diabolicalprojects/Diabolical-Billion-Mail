package video_outreach

import (
	"billionmail-core/internal/service/lead_scoring"
)

// TemplateSelection holds the chosen template ID and type for a contact.
type TemplateSelection struct {
	TemplateID int
	Type       string // "video" or "text" or "skip"
	Tier       string // "tier_1", "tier_2", or ""
}

// VideoOutreachConfig holds template IDs for each tier.
type VideoOutreachConfig struct {
	VideoTemplateID int // template ID for tier_1 (video email)
	TextTemplateID  int // template ID for tier_2 (text email)
}

// SelectTemplate picks the right template based on the contact's lead tier.
// Returns the template selection with type and tier info.
func SelectTemplate(cfg VideoOutreachConfig, attribs map[string]string) TemplateSelection {
	tier := attribs[lead_scoring.AttrLeadTier]

	switch tier {
	case lead_scoring.TagTier1:
		return TemplateSelection{
			TemplateID: cfg.VideoTemplateID,
			Type:       "video",
			Tier:       tier,
		}
	case lead_scoring.TagTier2:
		return TemplateSelection{
			TemplateID: cfg.TextTemplateID,
			Type:       "text",
			Tier:       tier,
		}
	default:
		return TemplateSelection{
			Type: "skip",
		}
	}
}

// VideoAttribKeys are the Contact.Attribs keys set by the video pipeline.
const (
	AttrVideoURL       = "video_url"
	AttrThumbnailURL   = "thumbnail_url"
	AttrLandingPageURL = "landing_page_url"
)

// HasVideoAssets checks if a contact has all required video assets in attribs.
func HasVideoAssets(attribs map[string]string) bool {
	return attribs[AttrVideoURL] != "" &&
		attribs[AttrThumbnailURL] != "" &&
		attribs[AttrLandingPageURL] != ""
}

// IsVideoEligible checks if a contact is tier_1 and has video assets ready.
func IsVideoEligible(attribs map[string]string) bool {
	return attribs[lead_scoring.AttrLeadTier] == lead_scoring.TagTier1 && HasVideoAssets(attribs)
}
