package filters

type VariantLength struct {
	LessThan1kbSingleGene     bool `yaml:"less_than_1kb_single_gene"`
	GreatThan1kbSingleGene    bool `yaml:"great_than_1kb_single_gene"`
	GreatThan1kbMultipleGenes bool `yaml:"great_than_1kb_multiple_genes"`
}

func init() {
	addFilter("VariantLength", new(VariantLength))
}

func (v *VariantLength) getSearchString() map[string]string {
	return map[string]string{
		"LessThan1kbSingleGene":     "1[VARLEN]:1000[VARLEN] AND \"single gene\"[Properties]",
		"GreatThan1kbSingleGene":    "1001[VARLEN]:300000000[VARLEN] AND \"single gene\"[Properties]",
		"GreatThan1kbMultipleGenes": "1001[VARLEN]:300000000[VARLEN] AND \"spans multiple genes\"[Properties]",
	}
}

func (v *VariantLength) getFilters() map[string]string {
	return map[string]string{
		"LessThan1kbSingleGene":     "< 1kb, single gene",
		"GreatThan1kbSingleGene":    "> 1kb, single gene",
		"GreatThan1kbMultipleGenes": "> 1kb, multiple genes",
	}
}

func (v *VariantLength) CreateQueryStringWithFilters(name []string) string {
	return buildSearchString(name, func(s string) string {
		return v.getSearchString()[s]
	})
}

func (v *VariantLength) PrintFilters(name []string) string {
	return buildFiltersString(name, func(s string) string {
		return v.getFilters()[s]
	})
}
