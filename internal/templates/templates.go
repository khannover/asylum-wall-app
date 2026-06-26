package templates

type Template struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	IncidentType   string `json:"incident_type"`
	ReasonCategory string `json:"reason_category"`
	Count          int    `json:"count"`
}

var All = []Template{
	{
		ID:           "platform_ban",
		Title:        "Banned from platform",
		Description:  "Account permanently or indefinitely banned.",
		IncidentType: "Account Ban",
	},
	{
		ID:           "platform_restriction",
		Title:        "Restricted on platform",
		Description:  "Features, uploads, or reach limited without a full ban.",
		IncidentType: "Reach Restriction",
	},
	{
		ID:           "shadowban",
		Title:        "Shadowbanned",
		Description:  "Content hidden or suppressed with no clear notice.",
		IncidentType: "Shadowban",
	},
	{
		ID:           "false_copyright",
		Title:        "False copyright strike",
		Description:  "Strike or claim on content you own or licensed.",
		IncidentType: "False Copyright Strike",
	},
	{
		ID:           "ai_flagged",
		Title:        "Flagged as AI / synthetic",
		Description:  "Content flagged as AI-generated or inauthentic.",
		IncidentType: "Reach Restriction",
		ReasonCategory: "AI-Detection False Positive",
	},
	{
		ID:           "fake_streams",
		Title:        "Accused of fake streaming",
		Description:  "Royalties withheld or penalized for alleged bot plays.",
		IncidentType: "Reach Restriction",
		ReasonCategory: "Alleged Artificial Streaming",
	},
	{
		ID:           "distribution_blocked",
		Title:        "Release blocked",
		Description:  "Distributor or platform rejected your release.",
		IncidentType: "Distribution Rejection",
	},
	{
		ID:           "content_removed",
		Title:        "Content removed",
		Description:  "Tracks, videos, or posts removed without clear reason.",
		IncidentType: "Reach Restriction",
		ReasonCategory: "Manual Abuse",
	},
	{
		ID:           "monetization_stripped",
		Title:        "Monetization disabled",
		Description:  "Ads, payouts, or creator fund access removed.",
		IncidentType: "Reach Restriction",
	},
	{
		ID:           "appeal_ignored",
		Title:        "Appeal ignored",
		Description:  "Support ticket or appeal went unanswered or denied.",
		IncidentType: "Account Ban",
		ReasonCategory: "Manual Abuse",
	},
}

func ByID(id string) (Template, bool) {
	for _, t := range All {
		if t.ID == id {
			return t, true
		}
	}
	return Template{}, false
}

func WithCounts(counts map[string]int) []Template {
	out := make([]Template, len(All))
	for i, t := range All {
		out[i] = t
		out[i].Count = counts[t.ID]
	}
	return out
}