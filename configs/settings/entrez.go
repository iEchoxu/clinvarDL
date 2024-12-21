package settings

import (
	"bufio"
	"github.com/iEchoxu/clinvarDL/configs/pkg/survey"
	"github.com/iEchoxu/clinvarDL/pkg/cdlerror"
	"os"

	"github.com/pkg/errors"
)

const (
	entrezSettingsToolName = "clinvarDL"
)

type EntrezSettings struct {
	DB         string `yaml:"db" `
	RetMax     int    `yaml:"ret_max" `
	RetMode    string `yaml:"ret_mode" `
	UseHistory bool   `yaml:"use_history" `
	SearchType string `yaml:"search_type"` // https://www.ncbi.nlm.nih.gov/clinvar/docs/help/
	Email      string `yaml:"email"`
	ToolName   string `yaml:"tool_name"`
	ApiKey     string `yaml:"api_key"`
	BatchSize  int    `yaml:"batch_size"`
}

func NewEntrezSettings() *EntrezSettings {
	return &EntrezSettings{
		DB:         "clinvar",
		RetMax:     10000,
		RetMode:    "xml",
		UseHistory: true,
		SearchType: "gene symbol",
		Email:      "",
		ToolName:   entrezSettingsToolName,
		ApiKey:     "",
		BatchSize:  10,
	}
}

func (es *EntrezSettings) InputProcessor() *EntrezSettings {
	reader := bufio.NewReader(os.Stdin)

	s := survey.Survey{
		Questions: []*survey.Question{
			{
				Text:     "请填写 Email,以便在 Entrez 出现问题时收到 NCBI 通知：",
				Answer:   "",
				Validate: survey.IsEmailValid,
			},
			{
				Text:     "是否使用 API_Key 以提升查询速度？（是请输入'y'，否请输入'n'）, 获取 API_Key 请访问: https://ncbiinsights.ncbi.nlm.nih.gov/2017/11/02/new-api-keys-for-the-e-utilities/：",
				Answer:   "",
				Validate: survey.IsYesNoValid,
				StopOnNo: true,
			},
			{
				Text:     "请输入 API Key: ",
				Answer:   "",
				Validate: survey.IsAPIKeyValid,
			},
		},
	}

	// 显示所有问题并获取答案
	s.AskAll(reader)

	es.Email = s.Questions[0].Answer

	apiKey := ""
	if s.Questions[1].Answer == "y" {
		apiKey = s.Questions[2].Answer
	}
	es.ApiKey = apiKey

	return es
}

func (es *EntrezSettings) SetEmailOrAPIKey(k, v string) error {
	switch k {
	case "email":
		if !survey.IsEmailValid(v) {
			return errors.Wrapf(cdlerror.ErrInvalidEmailAddress, "%s is not a valid email address", v)
		}
		es.Email = v
	case "apiKey":
		if !survey.IsAPIKeyValid(v) {
			return errors.Wrapf(cdlerror.ErrInvalidAPIKey, "%s is not a valid NCBI API key", v)
		}
		es.ApiKey = v
	default:
		return errors.Wrapf(cdlerror.ErrUnknownKey, "%s is not a valid key", v)
	}

	return nil
}
