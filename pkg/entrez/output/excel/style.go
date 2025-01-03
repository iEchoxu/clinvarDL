package excel

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

const (
	minStyle StyleType = iota
	AlternatingRow
	maxStyle // 用于边界检查
)

// styleRegistry 存储所有已注册的样式
var styleRegistry = make(map[StyleType]func() ExcelStyle)

type ExcelStyle interface {
	InitStyle(f *excelize.File) error
	GetRowStyle(rowCount int) int
}

type StyleType int

// init 初始化默认样式
func init() {
	RegisterStyle(AlternatingRow, func() ExcelStyle { return NewAlternatingRowStyle() })
}

// RegisterStyle 注册新的样式
func RegisterStyle(styleType StyleType, creator func() ExcelStyle) error {
	if styleType < minStyle || styleType >= maxStyle {
		return fmt.Errorf("invalid style type: %d", styleType)
	}
	styleRegistry[styleType] = creator
	return nil
}

// NewStyle 创建新的样式
func NewStyle(styleType StyleType) (ExcelStyle, error) {
	creator, ok := styleRegistry[styleType]
	if !ok {
		// 如果请求的样式类型不存在，返回默认样式
		creator = styleRegistry[AlternatingRow]
	}

	if creator == nil {
		return nil, fmt.Errorf("no style registered for type %d", styleType)
	}

	return creator(), nil
}

// AlternatingRowStyle 实现了交替行样式
type AlternatingRowStyle struct {
	headerStyle  int
	oddRowStyle  int
	evenRowStyle int
}

func NewAlternatingRowStyle() *AlternatingRowStyle {
	style := &AlternatingRowStyle{
		headerStyle:  0,
		oddRowStyle:  0,
		evenRowStyle: 0,
	}

	return style
}

func (ars *AlternatingRowStyle) InitStyle(f *excelize.File) error {
	// 创建并保存所有样式
	headerStyle, err := ars.setHeaderStyle(f)
	if err != nil {
		return err
	}

	oddRowStyle, err := ars.setOddRowStyle(f)
	if err != nil {
		return err
	}

	evenRowStyle, err := ars.setEvenRowStyle(f)
	if err != nil {
		return err
	}

	ars.headerStyle = headerStyle
	ars.oddRowStyle = oddRowStyle
	ars.evenRowStyle = evenRowStyle

	return nil
}

// GetRowStyle 根据行号返回对应的样式ID
func (ars *AlternatingRowStyle) GetRowStyle(rowCount int) int {
	if rowCount == 1 {
		return ars.headerStyle
	}
	if rowCount%2 == 0 {
		return ars.evenRowStyle
	}
	return ars.oddRowStyle
}

// setHeaderStyle 设置表头样式
func (ars *AlternatingRowStyle) setHeaderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11, Bold: true, Color: "#FFFFFF"},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472c4"},
			Pattern: 1,
		},
	})
}

// setOddRowStyle 设置奇数行样式
func (ars *AlternatingRowStyle) setOddRowStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#d9e1f2"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "8ea9db", Style: 1},
		},
	})
}

// setEvenRowStyle 设置偶数行样式
func (ars *AlternatingRowStyle) setEvenRowStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "e0e0e0", Style: 1},
		},
	})
}
