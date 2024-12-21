package excel

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/output"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"

	"github.com/xuri/excelize/v2"
)

type Writer struct {
	file         *excelize.File
	sheetName    string
	currentRow   int
	streamWriter *excelize.StreamWriter
	rowBuffer    [][]interface{}
	bufferSize   int
	mu           sync.Mutex
}

func NewWriter(sheetName string) (output.Writer, error) {
	f := excelize.NewFile()
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create new sheet: %w", err)
	}
	f.DeleteSheet("Sheet1") // 删除默认的Sheet1
	f.SetActiveSheet(index)

	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream writer: %w", err)
	}

	return &Writer{
		file:         f,
		sheetName:    sheetName,
		currentRow:   1,
		streamWriter: sw,
		rowBuffer:    make([][]interface{}, 0, 1000),
		bufferSize:   1000,
	}, nil
}

func (ew *Writer) SetHeaders(headers []string) error {
	headers = []string{
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
		"Query", // 添加查询列, 用于数据校对 （可删除）
	}

	ew.mu.Lock()
	defer ew.mu.Unlock()

	cell, _ := excelize.CoordinatesToCellName(1, ew.currentRow)

	interfaceHeaders := make([]interface{}, len(headers))
	for i, v := range headers {
		interfaceHeaders[i] = v
	}

	if err := ew.streamWriter.SetRow(cell, interfaceHeaders); err != nil {
		return fmt.Errorf("failed to set headers: %w", err)
	}
	ew.currentRow++
	return nil
}

func (ew *Writer) WriteResultStream(ctx context.Context, results <-chan *types.QueryResult) error {
	for {
		select {
		case result, ok := <-results:
			if !ok {
				return ew.flushBuffer()
			}

			rows, err := ew.processResult(result)
			if err != nil {
				return fmt.Errorf("failed to process result: %w", err)
			}

			ew.mu.Lock()
			ew.rowBuffer = append(ew.rowBuffer, rows...)
			if len(ew.rowBuffer) >= ew.bufferSize {
				if err := ew.flushBuffer(); err != nil {
					ew.mu.Unlock()
					return err
				}
			}
			ew.mu.Unlock()

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ew *Writer) processResult(result *types.QueryResult) ([][]interface{}, error) {
	if result == nil {
		return nil, nil
	}

	docSum := result.Result

	var rows [][]interface{}
	for _, doc := range docSum.DocumentSummarySet.DocumentSummary {
		// 构建 dbSNP ID
		var dbSNPIds []string
		for _, xref := range doc.VariationSet.Variation.VariationXrefs.VariationXref {
			if xref.DBSource == "dbSNP" {
				dbSNPIds = append(dbSNPIds, "rs"+xref.DbId)
			}
		}

		// 获取染色体位置信息
		var grch37Chr, grch37Loc, grch37Ver, grch38Chr, grch38Loc, grch38Ver string
		for _, assembly := range doc.VariationSet.Variation.VariationLoc.AssemblySet {
			if assembly.AssemblyName == "GRCh37" {
				grch37Chr = assembly.Chr
				grch37Ver = assembly.AssemblyAccVer
				if assembly.Start != "" && assembly.Stop != "" && assembly.Start != assembly.Stop {
					grch37Loc = fmt.Sprintf("%s-%s", assembly.Start, assembly.Stop)
				} else {
					grch37Loc = assembly.Start
				}
			} else if assembly.AssemblyName == "GRCh38" {
				grch38Chr = assembly.Chr
				grch38Ver = assembly.AssemblyAccVer
				if assembly.Start != "" && assembly.Stop != "" && assembly.Start != assembly.Stop {
					grch38Loc = fmt.Sprintf("%s-%s", assembly.Start, assembly.Stop)
				} else {
					grch38Loc = assembly.Start
				}
			}
		}

		// 构建条件列表
		var conditions []string
		for _, trait := range doc.GermlineClassification.TraitSet.Trait {
			conditions = append(conditions, trait.Name)
		}

		// 构建基因列表
		var genes, geneIDs []string
		for _, gene := range doc.Genes.Gene {
			genes = append(genes, gene.Symbol)
			geneIDs = append(geneIDs, gene.GeneID)
		}

		row := make([]interface{}, 30)
		row[0] = doc.Title // Name 字段改为 Title
		row[1] = strings.Join(genes, "|")
		row[2] = strings.Join(geneIDs, "|")
		row[3] = doc.ProteinChange
		row[4] = strings.Join(conditions, "|")
		row[5] = doc.Accession
		row[6] = doc.AccessionVersion
		row[7] = grch37Chr
		row[8] = grch37Loc
		row[9] = grch37Ver
		row[10] = grch38Chr
		row[11] = grch38Loc
		row[12] = grch38Ver
		row[13] = doc.Uid                              // VariationID 使用 Uid
		row[14] = doc.VariationSet.Variation.MeasureId // AlleleID(s) 使用 MeasureId
		row[15] = strings.Join(dbSNPIds, "|")          // dbSNP ID 从 VariationXrefs 构建
		row[16] = doc.VariationSet.Variation.CdnaChange
		row[17] = doc.VariationSet.Variation.CanonicalSPDI
		row[18] = doc.VariationSet.Variation.VariantType
		row[19] = strings.Join(doc.MolecularConsequenceList.String, "|") // 使用 String 而不是 Consequences
		row[20] = doc.GermlineClassification.Description
		row[21] = doc.GermlineClassification.LastEvaluated
		row[22] = doc.GermlineClassification.ReviewStatus
		row[23] = doc.ClinicalImpactClassification.Description
		row[24] = doc.ClinicalImpactClassification.LastEvaluated
		row[25] = doc.ClinicalImpactClassification.ReviewStatus
		row[26] = doc.OncogenicityClassification.Description
		row[27] = doc.OncogenicityClassification.LastEvaluated
		row[28] = doc.OncogenicityClassification.ReviewStatus
		row[29] = result.Query // 可删除

		rows = append(rows, row)
	}

	return rows, nil
}

func (ew *Writer) Save(filename string) error {
	ew.mu.Lock()
	defer ew.mu.Unlock()

	// 保存前确保缓冲区被刷新
	if len(ew.rowBuffer) > 0 {
		if err := ew.flushBuffer(); err != nil {
			return err
		}
	}

	if err := ew.streamWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush stream writer: %w", err)
	}
	return ew.file.SaveAs(filename)
}

func (ew *Writer) Close() error {
	return ew.file.Close()
}

func (ew *Writer) flushBuffer() error {
	for _, row := range ew.rowBuffer {
		cell, _ := excelize.CoordinatesToCellName(1, ew.currentRow)
		if err := ew.streamWriter.SetRow(cell, row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
		ew.currentRow++
	}
	ew.rowBuffer = ew.rowBuffer[:0] // 清空缓冲区但保留容量
	return nil
}
