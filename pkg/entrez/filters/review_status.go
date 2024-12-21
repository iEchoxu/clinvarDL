package filters

type ReviewStatus struct {
	PracticeGuideline          bool `yaml:"practice_guideline"`
	ExpertPanel                bool `yaml:"expert_panel"`
	MultipleSubmitters         bool `yaml:"multiple_submitters"`
	SingleSubmitter            bool `yaml:"single_submitter"`
	AtLeastOneStar             bool `yaml:"at_least_one_star"`
	ConflictingClassifications bool `yaml:"conflicting_classifications"`
}

func init() {
	addFilter("ReviewStatus", new(ReviewStatus))
}

func (r *ReviewStatus) getSearchString() map[string]string {
	return map[string]string{
		"PracticeGuideline":          "\"practice guideline\"[Review status]",
		"ExpertPanel":                "\"reviewed by expert panel\"[Review status]",
		"MultipleSubmitters":         "(\"criteria provided, multiple submitters, no conflicts\"[Review status])",
		"SingleSubmitter":            "\"criteria provided, single submitter\"[Review status]",
		"AtLeastOneStar":             "(\"criteria provided, conflicting classifications\"[Review status] OR \"criteria provided, multiple submitters, no conflicts\"[Review status] OR \"criteria provided, single submitter\"[Review status] OR \"practice guideline\"[Review status] OR \"reviewed by expert panel\"[Review status])",
		"ConflictingClassifications": "\"criteria provided, conflicting classifications\"[Review status]",
	}
}

func (r *ReviewStatus) getFilters() map[string]string {
	return map[string]string{
		"PracticeGuideline":          "Practice guideline",
		"ExpertPanel":                "Expert panel",
		"MultipleSubmitters":         "Multiple submitters",
		"SingleSubmitter":            "Single submitter",
		"AtLeastOneStar":             "At least one star",
		"ConflictingClassifications": "Conflicting classifications",
	}
}

func (r *ReviewStatus) CreateQueryStringWithFilters(name []string) string {
	return buildSearchString(name, func(s string) string {
		return r.getSearchString()[s]
	})
}

func (r *ReviewStatus) PrintFilters(name []string) string {
	return buildFiltersString(name, func(s string) string {
		return r.getFilters()[s]
	})
}
