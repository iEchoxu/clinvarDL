package filters

import (
	"errors"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"log"
	"strings"
)

var (
	filterMap = make(map[string]Filters, 20)
)

type Filters interface {
	CreateQueryStringWithFilters(name []string) string
	PrintFilters(name []string) string
}

func addFilter(name string, filter Filters) {
	if _, exists := filterMap[name]; exists {
		log.Fatalln(name, "filter already registered")
	}

	filterMap[name] = filter
}

func Get(name string) (Filters, error) {
	if _, ok := filterMap[name]; !ok {
		return nil, errors.New("没找到对应的 filter")
	}
	return filterMap[name], nil
}

// buildSearchString 函数接受一个字符串切片和一个处理函数，构建查询字符串
// 同一查询类型下的不同属性值会以逻辑或的方式组合起来（添加 OR 进行分割）
func buildSearchString(attrs []string, processFunc func(string) string) string {
	var builder strings.Builder

	// 如果只有一个属性值，则直接返回处理后的值且前后不用添加 ()
	if len(attrs) == 1 {
		return processFunc(attrs[0])
	}

	builder.WriteString("(")
	for i, attr := range attrs {
		builder.WriteString(processFunc(attr))
		if i < len(attrs)-1 {
			builder.WriteString(" OR ")
		}
	}
	builder.WriteString(")")

	return builder.String()
}

func buildFiltersString(attrs []string, processFunc func(string) string) string {
	var builder strings.Builder

	// 如果只有一个属性值，则直接返回处理后的值
	if len(attrs) == 1 {
		builder.WriteString(processFunc(attrs[0]))
		return builder.String()
	}

	for i, attr := range attrs {
		builder.WriteString(processFunc(attr))
		if i < len(attrs)-1 {
			builder.WriteString(",") // 根据需要修改分隔符,如 |
		}
	}

	return builder.String()
}

// BuildQueryStringWithTerm 构建 url 链接中的 Term 查询参数里的过滤条件
// 不同查询类型之间用 AND 进行分割
func BuildQueryStringWithTerm(activatedFilters map[string][]string) string {
	if len(activatedFilters) == 0 || activatedFilters == nil {
		return ""
	}

	var stringBuilder strings.Builder
	var filtersString strings.Builder
	stringBuilder.WriteString("(")
	count := 0

	for structName, attrList := range activatedFilters {
		filterIns, err := Get(structName)
		if err != nil {
			log.Println(err)
		}

		filtersString.WriteString(filterIns.PrintFilters(attrList)) // 输出条件过滤器信息

		stringBuilder.WriteString(filterIns.CreateQueryStringWithFilters(attrList)) // 构建 term 中的查询条件字符串

		if count < len(activatedFilters)-1 {
			stringBuilder.WriteString(" AND ")
			filtersString.WriteString(" | ")
		}
		count++
	}

	stringBuilder.WriteString(")")

	logcdl.Tip("filters activated: %s\n", filtersString.String())

	return stringBuilder.String()
}
