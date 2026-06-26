package entries

func TemplateCounts(list []Entry) map[string]int {
	counts := make(map[string]int)
	for _, e := range list {
		if e.TemplateID != "" {
			counts[e.TemplateID]++
		}
	}
	return counts
}