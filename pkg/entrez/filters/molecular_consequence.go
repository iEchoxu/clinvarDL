package filters

type MolecularConsequence struct {
	Frameshift bool `yaml:"frameshift"`
	Missense   bool `yaml:"missense"`
	Nonsense   bool `yaml:"nonsense"`
	SpliceSite bool `yaml:"splice_site"`
	NcRNA      bool `yaml:"ncRNA"`
	NearGene   bool `yaml:"near_gene"`
	Utr        bool `yaml:"utr"`
}

func init() {

	addFilter("MolecularConsequence", new(MolecularConsequence))
}

func (m *MolecularConsequence) getSearchString() map[string]string {
	return map[string]string{
		"Frameshift": "\"frameshift variant\"[molecular consequence]",
		"Missense":   "(\"missense variant\"[molecular consequence] OR \"SO 0001583\"[molecular consequence])",
		"Nonsense":   "(\"nonsense\"[molecular consequence] OR \"SO 0001587\"[molecular consequence])",
		"SpliceSite": "(\"splice 3\"[Molecular consequence] OR \"splice 5\"[Molecular consequence] OR \"splice site\"[Molecular consequence] OR \"splice donor variant\"[Molecular consequence] OR \"splice acceptor variant\"[molecular consequence])",
		"NcRNA":      "\"non coding transcript variant\"[molecular consequence]",
		"NearGene":   "(\"500b downstream variant\"[molecular consequence] OR \"2kb upstream variant\"[molecular consequence])",
		"Utr":        "(\"3 prime utr variant\"[molecular consequence] OR \"5 prime utr variant\"[molecular consequence])",
	}
}

func (m *MolecularConsequence) getFilters() map[string]string {
	return map[string]string{
		"Frameshift": "Frameshift",
		"Missense":   "Missense",
		"Nonsense":   "Nonsense",
		"SpliceSite": "Splice site",
		"NcRNA":      "ncRNA",
		"NearGene":   "Near gene",
		"Utr":        "UTR",
	}
}

func (m *MolecularConsequence) CreateQueryStringWithFilters(name []string) string {
	return buildSearchString(name, func(s string) string {
		return m.getSearchString()[s]
	})
}

func (m *MolecularConsequence) PrintFilters(name []string) string {
	return buildFiltersString(name, func(s string) string {
		return m.getFilters()[s]
	})
}
