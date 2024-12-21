package configs

import (
	"github.com/iEchoxu/clinvarDL/pkg/entrez/filters"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"reflect"
)

type FiltersConfig struct {
	Path                   string                         `yaml:"-"`
	ClassificationType     filters.ClassificationType     `yaml:"classification_type"`
	GermlineClassification filters.GermlineClassification `yaml:"germline_classification"`
	TypesOfConflicts       filters.TypesOfConflicts       `yaml:"types_of_conflicts"`
	MolecularConsequence   filters.MolecularConsequence   `yaml:"molecular_consequence"`
	VariationType          filters.VariationType          `yaml:"variation_type"`
	VariationSize          filters.VariationSize          `yaml:"variation_size"`
	VariantLength          filters.VariantLength          `yaml:"variant_length"`
	ReviewStatus           filters.ReviewStatus           `yaml:"review_status"`
}

// NewFiltersConfig 创建并返回一个FiltersConfig配置，不带参数
// 可用于初始化 configs.Config{}
func NewFiltersConfig() *FiltersConfig {
	return NewFiltersConfigWithPath("")
}

// NewFiltersConfigWithPath 创建并返回一个FiltersConfig配置，带一个path参数
// 在使用 BuildQueryStringWithTerm 构建 url 链接中的 Term 查询参数时可用
func NewFiltersConfigWithPath(path string) *FiltersConfig {
	// 设置默认值
	defaults := &FiltersConfig{
		Path:                   "",
		ClassificationType:     filters.ClassificationType{},
		GermlineClassification: filters.GermlineClassification{},
		TypesOfConflicts:       filters.TypesOfConflicts{},
		MolecularConsequence:   filters.MolecularConsequence{},
		VariationType:          filters.VariationType{},
		VariationSize:          filters.VariationSize{},
		VariantLength:          filters.VariantLength{},
		ReviewStatus:           filters.ReviewStatus{},
	}

	if path != "" {
		defaults.Path = path
	}

	return defaults
}

func (fc *FiltersConfig) Write(configFile string) error {
	return write(fc, configFile)
}

func (fc *FiltersConfig) Read(configFile string) (Configer, error) {
	return read(configFile, fc)
}

// getActivatedFilter 获得 filters 配置文件中值为 true 的属性
func (fc *FiltersConfig) getActivatedFilter(configFile string) map[string][]string {
	if configFile == "" {
		logcdl.Warn("filters config file is empty")
		return nil
	}

	configer, err := fc.Read(configFile)
	if err != nil {
		logcdl.Error("failed to read filters config: %v", err)
		return nil
	}

	activatedFilters := make(map[string][]string, 30) // 如果 filters 数据增多可适当增加此数值

	// 只有结构体才能使用反射，指针类型的需要解引用
	t := reflect.TypeOf(*configer.(*FiltersConfig))
	v := reflect.ValueOf(*configer.(*FiltersConfig))

	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i).Type
		if fieldType.Kind() == reflect.Struct {
			for j := 0; j < fieldType.NumField(); j++ {
				subFieldValue := fieldValue.Field(j)
				if subFieldValue.IsValid() && subFieldValue.Bool() {
					activatedFilters[t.Field(i).Name] = append(activatedFilters[t.Field(i).Name], fieldType.Field(j).Name)
				}
			}
		}
	}

	return activatedFilters
}

// BuildQueryStringWithTerm 构建 url 链接中的 Term 查询参数里的过滤条件
func (fc *FiltersConfig) BuildQueryStringWithTerm() string {
	return filters.BuildQueryStringWithTerm(fc.getActivatedFilter(fc.Path))
}
