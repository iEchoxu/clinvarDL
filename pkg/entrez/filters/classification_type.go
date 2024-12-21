package filters

type ClassificationType struct {
	Germline bool `yaml:"germline"`
	Somatic  bool `yaml:"somatic"`
}

func init() {
	addFilter("ClassificationType", new(ClassificationType))
}

// getSearchString 设置默认的查询语句，在 https://www.ncbi.nlm.nih.gov/clinvar/advanced 获取
func (c *ClassificationType) getSearchString() map[string]string {
	return map[string]string{
		"Germline": "\"germline_classification\"[PROP]",
		"Somatic":  "(\"somatic_clinical_impact_classification\"[PROP] OR \"oncogenicity_classification\"[PROP])",
	}
}

func (c *ClassificationType) getFilters() map[string]string {
	return map[string]string{
		"Germline": "Germline",
		"Somatic":  "Somatic",
	}
}

// CreateQueryStringWithFilters 创建查询字符串中的查询条件筛选项,返回 string
func (c *ClassificationType) CreateQueryStringWithFilters(name []string) string {
	return buildSearchString(name, func(s string) string {
		return c.getSearchString()[s]
	})
}

func (c *ClassificationType) PrintFilters(name []string) string {
	return buildFiltersString(name, func(s string) string {
		return c.getFilters()[s]
	})
}
