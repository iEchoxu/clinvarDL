package filters

type GermlineClassification struct {
	ConflictingClassifications bool `yaml:"conflicting_classifications"`
	Benign                     bool `yaml:"benign"`
	LikelyBenign               bool `yaml:"likely_benign"`
	UncertainSignificance      bool `yaml:"uncertain_significance"`
	LikelyPathogenic           bool `yaml:"likely_pathogenic"`
	Pathogenic                 bool `yaml:"pathogenic"`
}

func init() {
	addFilter("GermlineClassification", new(GermlineClassification))
}

func (g *GermlineClassification) getSearchString() map[string]string {
	return map[string]string{
		"ConflictingClassifications": "\"clinsig has conflicts\"[Properties]",
		"Benign":                     "\"clinsig benign\"[Properties]",
		"LikelyBenign":               "\"clinsig likely benign\"[Properties]",
		"UncertainSignificance":      "(\"clinsig vus\"[Properties] or \"clinsig uncertain risk allele\"[Properties])",
		"LikelyPathogenic":           "(\"clinsig likely pathogenic\"[Properties] or \"clinsig likely pathogenic low penetrance\"[Properties] or \"clinsig likely risk allele\"[Properties])",
		"Pathogenic":                 "(\"clinsig pathogenic\"[Properties] or \"clinsig pathogenic low penetrance\"[Properties] or \"clinsig established risk allele\"[Properties])",
	}
}

func (g *GermlineClassification) getFilters() map[string]string {
	return map[string]string{
		"ConflictingClassifications": "Conflicting classifications",
		"Benign":                     "Benign",
		"LikelyBenign":               "Likely benign",
		"UncertainSignificance":      "Uncertain significance",
		"LikelyPathogenic":           "Likely pathogenic",
		"Pathogenic":                 "Pathogenic",
	}
}

func (g *GermlineClassification) CreateQueryStringWithFilters(name []string) string {
	return buildSearchString(name, func(s string) string {
		return g.getSearchString()[s]
	})
}

func (g *GermlineClassification) PrintFilters(name []string) string {
	return buildFiltersString(name, func(s string) string {
		return g.getFilters()[s]
	})
}
