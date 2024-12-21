package filters

type VariationSize struct {
	ShortVariantLessThan50bps       bool `yaml:"short_variant_less_than_50bps"`
	StructuralVariantGreatThan50bps bool `yaml:"structural_variant_great_than_50bps"`
}

func init() {
	addFilter("VariationSize", new(VariationSize))
}

func (vs *VariationSize) getSearchString() map[string]string {
	return map[string]string{
		"ShortVariantLessThan50bps":       "0[VARLEN]:49[VARLEN]",
		"StructuralVariantGreatThan50bps": "50[VARLEN]:2000000000[VARLEN]",
	}
}

func (vs *VariationSize) getFilters() map[string]string {
	return map[string]string{
		"ShortVariantLessThan50bps":       "Short variant (< 50 bps)",
		"StructuralVariantGreatThan50bps": "Structural variant (>= 50 bps)",
	}
}

func (vs *VariationSize) CreateQueryStringWithFilters(name []string) string {
	return buildSearchString(name, func(s string) string {
		return vs.getSearchString()[s]
	})
}

func (vs *VariationSize) PrintFilters(name []string) string {
	return buildFiltersString(name, func(s string) string {
		return vs.getFilters()[s]
	})
}
