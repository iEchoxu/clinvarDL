package types

// ESearchResult 定义了 ESearch 操作的结果结构
type ESearchResult struct {
	Count  int `xml:"Count"`
	IdList struct {
		Id []string `xml:"Id"`
	} `xml:"IdList"`
	QueryKey string `xml:"QueryKey"`
	WebEnv   string `xml:"WebEnv"`
}

// EPostResult 定义了 EPost 操作的结果结构
type EPostResult struct {
	QueryKey string `xml:"QueryKey"`
	WebEnv   string `xml:"WebEnv"`
}

// ESummaryResult 定义了 ESummary 操作的结果结构
type ESummaryResult struct {
	DocumentSummarySet DocumentSummarySet `xml:"DocumentSummarySet"`
}

// DocumentSummarySet 定义了文档摘要集合
type DocumentSummarySet struct {
	DocumentSummary []*DocumentSummary `xml:"DocumentSummary"`
}

// DocumentSummary 定义了单个文档摘要
type DocumentSummary struct {
	Uid                          string          `xml:"uid,attr"`
	Accession                    string          `xml:"accession"`
	AccessionVersion             string          `xml:"accession_version"`
	Title                        string          `xml:"title"`
	VariationSet                 VariationSet    `xml:"variation_set"`
	GermlineClassification       Classification  `xml:"germline_classification"`
	ClinicalImpactClassification Classification  `xml:"clinical_impact_classification"`
	OncogenicityClassification   Classification  `xml:"oncogenicity_classification"`
	GeneSort                     string          `xml:"gene_sort"`
	ChrSort                      string          `xml:"chr_sort"`
	LocationSort                 string          `xml:"location_sort"`
	Genes                        GeneList        `xml:"genes"`
	MolecularConsequenceList     ConsequenceList `xml:"molecular_consequence_list"`
	ProteinChange                string          `xml:"protein_change"`
}

// VariationSet 定义了变异集合
type VariationSet struct {
	Variation Variation `xml:"variation"`
}

// Variation 定义了变异信息
type Variation struct {
	MeasureId      string         `xml:"measure_id"`
	VariationXrefs VariationXrefs `xml:"variation_xrefs"`
	CdnaChange     string         `xml:"cdna_change"`
	VariationLoc   VariationLoc   `xml:"variation_loc"`
	VariantType    string         `xml:"variant_type"`
	CanonicalSPDI  string         `xml:"canonical_spdi"`
}

// XrefList 定义了变异引用列表
type VariationXrefs struct {
	VariationXref []VariationXref `xml:"variation_xref"`
}

// VariationXref 定义了变异引用
type VariationXref struct {
	DBSource string `xml:"db_source"`
	DbId     string `xml:"db_id"`
}

// VariationLoc 定义了变异位置
type VariationLoc struct {
	AssemblySet []Assembly `xml:"assembly_set"`
}

// Assembly 定义了基因组装信息
type Assembly struct {
	Status            string `xml:"status"`
	AssemblyName      string `xml:"assembly_name"`
	Chr               string `xml:"chr"`
	Band              string `xml:"band"`
	Start             string `xml:"start"`
	Stop              string `xml:"stop"`
	DisplayStart      string `xml:"display_start"`
	DisplayStop       string `xml:"display_stop"`
	AssemblyAccVer    string `xml:"assembly_acc_ver"`
	AnnotationRelease string `xml:"annotation_release"`
}

// Classification 定义了分类信息
type Classification struct {
	Description   string   `xml:"description"`
	LastEvaluated string   `xml:"last_evaluated"`
	ReviewStatus  string   `xml:"review_status"`
	TraitSet      TraitSet `xml:"trait_set,omitempty"`
}

// TraitSet 定义了特征集合
type TraitSet struct {
	Trait []TraitInfo `xml:"trait"`
}

// GeneList 定义了基因列表
type GeneList struct {
	Gene []Gene `xml:"gene"`
}

// Gene 定义了基因信息
type Gene struct {
	Symbol string `xml:"symbol"`
	GeneID string `xml:"GeneID"`
}

// ConsequenceList 定义了分子后果列表
type ConsequenceList struct {
	String []string `xml:"string"`
}

// TraitInfo 定义了特征信息的结构
type TraitInfo struct {
	TraitXrefs struct {
		TraitXref []struct {
			DBSource string `xml:"db_source"`
			DbId     string `xml:"db_id"`
		} `xml:"trait_xref"`
	} `xml:"trait_xrefs"`
	Name string `xml:"trait_name"`
}
