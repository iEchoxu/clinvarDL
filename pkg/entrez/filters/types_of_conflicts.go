package filters

type TypesOfConflicts struct {
	PLPVsLBB bool `yaml:"plp_vs_lbb"`
	PLPVsVUS bool `yaml:"plp_vs_vus"`
	VUSVsLBB bool `yaml:"vus_vs_lbb"`
}

func init() {
	addFilter("TypesOfConflicts", new(TypesOfConflicts))
}

func (t *TypesOfConflicts) getSearchString() map[string]string {
	return map[string]string{
		"PLPVsLBB": "\"clinsig conf plp vs lbb\"[Properties]",
		"PLPVsVUS": "\"clinsig conf plp vs vus\"[Properties]",
		"VUSVsLBB": "\"clinsig conf vus vs lbb\"[Properties]",
	}
}

func (t *TypesOfConflicts) getFilters() map[string]string {
	return map[string]string{
		"PLPVsLBB": "P/LP vs LB/B",
		"PLPVsVUS": "P/LP vs VUS",
		"VUSVsLBB": "VUS vs LB/B",
	}
}

func (t *TypesOfConflicts) CreateQueryStringWithFilters(name []string) string {
	return buildSearchString(name, func(s string) string {
		return t.getSearchString()[s]
	})
}

func (t *TypesOfConflicts) PrintFilters(name []string) string {
	return buildFiltersString(name, func(s string) string {
		return t.getFilters()[s]
	})
}
