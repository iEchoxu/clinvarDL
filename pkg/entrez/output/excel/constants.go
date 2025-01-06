package excel

const (
	defaultRowHeight  = 25.0 // 设置默认行高为 25
	defaultColCount   = 29   // 默认列数
	activeStyle       = AlternatingRow
	defaultBufferSize = 1000 // 默认缓冲区大小
)

// 定义列宽常量
var defaultColumnWidths = [defaultColCount]float64{
	45, // A: Name
	20, // B: Genes
	20, // C: GeneIDs
	28, // D: Protein change
	62, // E: Conditions
	20, // F: Accession
	20, // G: Accession Version
	20, // H: GRCh37Chromosome
	26, // I: GRCh37Location
	26, // J: GRCh37AssemblyAccVer
	20, // K: GRCh38Chromosome
	20, // L: GRCh38Location
	20, // M: GRCh38AssemblyAccVer
	20, // N: VariationID
	20, // O: AlleleID(s)
	20, // P: dbSNP ID
	40, // Q: Cdna Change
	44, // R: Canonical SPDI
	28, // S: Variant type
	60, // T: Molecular consequence
	36, // U: Germline classification
	28, // V: Germline date last evaluated
	48, // W: Germline review status
	20, // X: Somatic clinical impact
	38, // Y: Somatic clinical impact date last evaluated
	36, // Z: Somatic clinical impact review status
	26, // AA: Oncogenicity classification
	32, // AB: Oncogenicity date last evaluated
	28, // AC: Oncogenicity review status
}

// 定义表头常量
var defaultHeaders = [defaultColCount]string{
	"Name",
	"Gene(s)",
	"GeneID",
	"Protein change",
	"Condition(s)",
	"Accession",
	"Accession Version",
	"GRCh37Chromosome",
	"GRCh37Location",
	"GRCh37AssemblyAccVer",
	"GRCh38Chromosome",
	"GRCh38Location",
	"GRCh38AssemblyAccVer",
	"VariationID",
	"AlleleID(s)",
	"dbSNP ID",
	"Cdna Change",
	"Canonical SPDI",
	"Variant type",
	"Molecular consequence",
	"Germline classification",
	"Germline date last evaluated",
	"Germline review status",
	"Somatic clinical impact",
	"Somatic clinical impact date last evaluated",
	"Somatic clinical impact review status",
	"Oncogenicity classification",
	"Oncogenicity date last evaluated",
	"Oncogenicity review status",
	// "Query", // 添加查询列，用于数据校对 （可删除）
}
