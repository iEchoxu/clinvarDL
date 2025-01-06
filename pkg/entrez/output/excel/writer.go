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
	currentRow   int
	streamWriter *excelize.StreamWriter
	rowBuffer    [][]interface{}
	mu           sync.Mutex
	styles       ExcelStyle
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

	// 获取当前激活的样式
	styles, err := NewStyle(activeStyle)
	if err != nil {
		return nil, fmt.Errorf("failed to create style: %w", err)
	}

	// 初始化样式
	if err := styles.InitStyle(f); err != nil {
		return nil, fmt.Errorf("failed to init style: %w", err)
	}

	w := &Writer{
		file:         f,
		currentRow:   1,
		streamWriter: sw,
		rowBuffer:    make([][]interface{}, 0, defaultBufferSize),
		styles:       styles,
	}

	return w, nil
}

func (ew *Writer) SetHeaders(headers []string) error {
	// 使用默认表头
	if headers == nil {
		headers = defaultHeaders[:]
	}

	// 使用预定义的列宽
	for i, width := range defaultColumnWidths {
		if err := ew.streamWriter.SetColWidth(i+1, i+1, width); err != nil {
			return fmt.Errorf("failed to set column width: %w", err)
		}
	}

	// 冻结首行
	if err := ew.streamWriter.SetPanes(&excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return fmt.Errorf("failed to set panes: %w", err)
	}

	// 写入表头
	cell, _ := excelize.CoordinatesToCellName(1, ew.currentRow)
	interfaceHeaders := make([]interface{}, len(headers))
	for i, v := range headers {
		interfaceHeaders[i] = v
	}

	if err := ew.streamWriter.SetRow(cell, interfaceHeaders, excelize.RowOpts{
		Height:  defaultRowHeight,
		StyleID: ew.styles.GetRowStyle(ew.currentRow), // 设置表头样式
	}); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
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

			// 写入缓冲区
			if len(rows) > 0 {
				ew.mu.Lock()
				ew.rowBuffer = append(ew.rowBuffer, rows...)
				needFlush := len(ew.rowBuffer) >= defaultBufferSize
				if needFlush {
					if err := ew.flushBuffer(); err != nil {
						ew.mu.Unlock()
						return err
					}
				}
				ew.mu.Unlock()
			}

		case <-ctx.Done():
			// 确保在上下文取消时也能刷新缓冲区
			if len(ew.rowBuffer) > 0 {
				if err := ew.flushBuffer(); err != nil {
					return fmt.Errorf("failed to flush buffer on context done: %w", err)
				}
			}
			return ctx.Err()
		}
	}
}

func (ew *Writer) processResult(result *types.QueryResult) ([][]interface{}, error) {
	if result == nil || result.Result == nil ||
		len(result.Result.DocumentSummarySet.DocumentSummary) == 0 {
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

		row := make([]interface{}, defaultColCount)
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
		// row[29] = result.Query // 可删除

		rows = append(rows, row)
	}

	return rows, nil
}

func (ew *Writer) Save(filename string) error {
	ew.mu.Lock()
	defer ew.mu.Unlock()

	// 刷新并关闭 StreamWriter
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

		if err := ew.streamWriter.SetRow(cell, row, excelize.RowOpts{
			Height:  defaultRowHeight,
			StyleID: ew.styles.GetRowStyle(ew.currentRow), // 使用自定义样式
		}); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
		ew.currentRow++
	}

	// 添加表格
	lastCol, _ := excelize.ColumnNumberToName(len(defaultHeaders))
	if err := ew.streamWriter.AddTable(&excelize.Table{
		Range: fmt.Sprintf("A1:%s%d", lastCol, ew.currentRow),
		Name:  "Table1",
		// StyleName: "TableStyleMedium13", // 使用内置样式
	}); err != nil {
		return fmt.Errorf("failed to add table: %w", err)
	}

	// 清空缓冲区
	ew.rowBuffer = ew.rowBuffer[:0]
	return nil
}
