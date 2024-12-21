package unuse

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Result struct {
	ResultItem map[string]json.RawMessage `json:"result"`
}

// ResultItem 代表单个变异信息的结构体
type ResultItem struct {
	Title                    string                 `json:"title"`
	Genes                    []Genes                `json:"genes"`
	ProteinChange            string                 `json:"protein_change"`
	GermlineClassification   GermlineClassification `json:"germline_classification"`
	Accession                string                 `json:"accession"`
	VariationSet             []VariationSet         `json:"variation_set"`
	Uid                      string                 `json:"uid"`
	MolecularConsequenceList []string               `json:"molecular_consequence_list"`
	Gene                     string                 // 新增字段: 用于存储解析到的 Genes 切片的值
	MolecularConsequences    string                 // 新增字段
}

// UnmarshalJSON 自定义解析 ResultItem
// 同时对 Genes、GermlineClassification 里的数据进行拼接并保存到 Gene、MolecularConsequences 中
func (res *ResultItem) UnmarshalJSON(bytes []byte) error {
	type resultItemAlias ResultItem // 避免递归调用 ResultItem 进入死循环
	var resultItem resultItemAlias
	if err := json.Unmarshal(bytes, &resultItem); err != nil {
		return err
	}

	var genesOfSymbol []string
	for _, v := range resultItem.Genes {
		genesOfSymbol = append(genesOfSymbol, v.Symbol)
	}

	geneRes := joinWithPipe(genesOfSymbol, "|")
	molecularConsequenceList := joinWithPipe(resultItem.MolecularConsequenceList, "|")

	*res = ResultItem(resultItem) // 将 resultItem 的值赋给 res 指向的 ResultItem 实例,可能有浅拷贝问题
	res.Gene = geneRes
	res.MolecularConsequences = molecularConsequenceList

	return nil
}

type Genes struct {
	Symbol string `json:"symbol"`
}

type GermlineClassification struct {
	Description   string     `json:"description"`
	LastEvaluated string     `json:"last_evaluated"`
	ReviewStatus  string     `json:"review_status"`
	TraitSet      []TraitSet `json:"trait_set"`
	TraitNames    string
}

// UnmarshalJSON 解析 []TraitSet 数据拼接并存储在 TraitNames 字段中
func (g *GermlineClassification) UnmarshalJSON(bytes []byte) error {
	type germlineClassificationAlias GermlineClassification
	var germlineClassification germlineClassificationAlias
	if err := json.Unmarshal(bytes, &germlineClassification); err != nil {
		return err
	}

	var traitSetArray []string
	for _, v := range germlineClassification.TraitSet {
		traitSetArray = append(traitSetArray, v.TraitName)
	}

	traitNameStr := joinWithPipe(traitSetArray, "|")

	g.TraitNames = traitNameStr
	g.Description = germlineClassification.Description
	g.LastEvaluated = germlineClassification.LastEvaluated
	g.ReviewStatus = germlineClassification.ReviewStatus

	return nil
}

type TraitSet struct {
	TraitName string `json:"trait_name"`
}

type VariationSet struct {
	MeasureId      string           `json:"measure_id"`
	VariationXrefs []VariationXrefs `json:"variation_xrefs"`
	VariationLoc   []VariationLoc   `json:"variation_loc"`
	VariantType    string           `json:"variant_type"`
	CanonicalSpdi  string           `json:"canonical_spdi"`
	// 下面的字段用于存储值，上面的用于解析 json
	dbSNPId       string
	VariationLocS map[string]string
}

// UnmarshalJSON 对 VariationXrefs 中解析到的数据添加 rs 前缀并存储在 dbSNPId
// 构建一个键为 : GRCh37Chromosome GRCh38Chromosome GRCh37Location GRCh38Location 的 Map 存储数据
func (vSet *VariationSet) UnmarshalJSON(bytes []byte) error {
	type variationSetAlias VariationSet
	var variationSet variationSetAlias
	if err := json.Unmarshal(bytes, &variationSet); err != nil {
		return err
	}

	var variationXrefsOfDbId []string
	for _, v := range variationSet.VariationXrefs {
		if v.DBSource == "dbSNP" {
			dbID := "rs" + v.DbId
			variationXrefsOfDbId = append(variationXrefsOfDbId, dbID)
		}
	}

	dbIDString := joinWithPipe(variationXrefsOfDbId, "|")

	vSet.dbSNPId = dbIDString
	vSet.MeasureId = variationSet.MeasureId
	vSet.VariantType = variationSet.VariantType
	vSet.CanonicalSpdi = variationSet.CanonicalSpdi

	// 初始化 Map: 如果 Start!=Stop 返回 "Start - Stop" 这样格式的数据
	// Map 键是 AssemblyName+"Chromosome" 以及 AssemblyName+"Location"
	// AssemblyName 值基本固定为: GRCh37 GRCh38 这两种
	vSet.VariationLocS = make(map[string]string)
	for _, j := range variationSet.VariationLoc {
		grchLocation := j.Start
		if j.Start != j.Stop {
			grchLocation = fmt.Sprintf("%s - %s", j.Start, j.Stop)
		}

		vSet.VariationLocS[j.AssemblyName+"Chromosome"] = j.Chr
		vSet.VariationLocS[j.AssemblyName+"Location"] = grchLocation
	}

	return nil
}

type VariationXrefs struct {
	DBSource string `json:"db_source"`
	DbId     string `json:"db_id"`
}

type VariationLoc struct {
	Status            string `json:"status"`
	AssemblyName      string `json:"assembly_name"`
	Chr               string `json:"chr"`
	Band              string `json:"band"`
	Start             string `json:"start"`
	Stop              string `json:"stop"`
	DisplayStart      string `json:"display_start"`
	DisplayStop       string `json:"display_stop"`
	AssemblyAccVer    string `json:"assembly_acc_ver"`
	AnnotationRelease string `json:"annotation_release"`
}

// ParaResult 只支持 uids <= 500 的请求
func ParaResult(bytes []byte) error {
	var result Result

	if err := json.Unmarshal(bytes, &result); err != nil {
		log.Fatalln("解析JSON出错:", err)
		return err
	}

	var ids []string
	err := json.Unmarshal(result.ResultItem["uids"], &ids)
	if err != nil {
		log.Println(err)
		return err
	}

	var resultItem ResultItem
	for _, id := range ids {
		err = json.Unmarshal(result.ResultItem[id], &resultItem)
		if err != nil {
			log.Println(err)
			//continue
			return err
		}
		DoParse(&resultItem)
	}

	return nil
}

func DoParse(resultItem *ResultItem) {
	fmt.Println(strings.Repeat("-", 120))

	fmt.Println("Name is:", resultItem.Title)
	fmt.Println("Gene(s) is:", resultItem.Gene)
	fmt.Println("Protein change is:", resultItem.ProteinChange) // 可以为空
	fmt.Println("Condition(s) is:", resultItem.GermlineClassification.TraitNames)
	fmt.Println("Accession is:", resultItem.Accession)
	for _, v := range resultItem.VariationSet {
		fmt.Println("GRCh37Chromosome is:", v.VariationLocS["GRCh37Chromosome"])
		fmt.Println("GRCh37Location is:", v.VariationLocS["GRCh37Location"])
		fmt.Println("GRCh38Chromosome is:", v.VariationLocS["GRCh38Chromosome"])
		fmt.Println("GRCh38Location is:", v.VariationLocS["GRCh38Location"])
		fmt.Println("AlleleID(s) is:", v.MeasureId)
		fmt.Println("dbSNP ID is: ", v.dbSNPId)
		fmt.Println("Canonical SPDI is:", v.CanonicalSpdi)
		fmt.Println("Variant type is:", v.VariantType)
	}

	fmt.Println("VariationID is:", resultItem.Uid)
	fmt.Println("Molecular consequence is:", resultItem.MolecularConsequences)
	fmt.Println("Germline classification is:", resultItem.GermlineClassification.Description)
	fmt.Println("Germline date last evaluated is:", resultItem.GermlineClassification.LastEvaluated)
	fmt.Println("Germline review status is:", resultItem.GermlineClassification.ReviewStatus)
}

// joinWithPipe 当切片中的数据超过 1 个时用 sep 进行拼接
func joinWithPipe(elements []string, sep string) string {
	if len(elements) == 0 {
		return ""
	}
	if len(elements) == 1 {
		return elements[0]
	}
	var builder strings.Builder
	builder.WriteString(elements[0])
	for _, elem := range elements[1:] {
		builder.WriteString(sep)
		builder.WriteString(elem)
	}
	return builder.String()
}
