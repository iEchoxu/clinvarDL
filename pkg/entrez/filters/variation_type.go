package filters

type VariationType struct {
	Deletion         bool `yaml:"deletion"`
	Duplication      bool `yaml:"duplication"`
	Indel            bool `yaml:"indel"`
	Insertion        bool `yaml:"insertion"`
	SingleNucleotide bool `yaml:"single_nucleotide"`
}

func init() {
	addFilter("VariationType", new(VariationType))
}

func (v *VariationType) getSearchString() map[string]string {
	return map[string]string{
		"Deletion":         "(\"deletion\"[Type of variation] OR \"copy number loss\"[Type of variation] or \"indel\"[Type of variation])",
		"Duplication":      "(\"duplication\"[Type of variation] OR \"copy number gain\"[Type of variation])",
		"Indel":            "\"indel\"[Type of variation]",
		"Insertion":        "(\"insertion\"[Type of variation] OR \"indel\"[Type of variation] OR \"duplication\"[Type of variation])",
		"SingleNucleotide": "\"single nucleotide variant\"[Type of variation]",
	}
}

func (v *VariationType) getFilters() map[string]string {
	return map[string]string{
		"Deletion":         "Deletion",
		"Duplication":      "Duplication",
		"Indel":            "Indel",
		"Insertion":        "Insertion",
		"SingleNucleotide": "Single nucleotide",
	}
}
func (v *VariationType) CreateQueryStringWithFilters(name []string) string {
	return buildSearchString(name, func(s string) string {
		return v.getSearchString()[s]
	})
}

func (v *VariationType) PrintFilters(name []string) string {
	return buildFiltersString(name, func(s string) string {
		return v.getFilters()[s]
	})
}
