package survey

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// Question 存储单个问题、答案和验证函数的结构体
type Question struct {
	Text     string
	Answer   string
	Validate func(string) bool
	StopOnNo bool
}

// Survey 管理多个问题的调查问卷结构体
type Survey struct {
	Questions []*Question // 包含多个问题的切片
}

// AskAll 显示Survey中的所有问题并获取答案
func (s *Survey) AskAll(reader *bufio.Reader) {
	for i, question := range s.Questions {
		for {
			fmt.Print(question.Text)
			question.Answer, _ = reader.ReadString('\n')
			question.Answer = strings.TrimSpace(question.Answer) // 去除换行符

			// 如果答案无效，提示用户重新输入
			if question.Validate(question.Answer) {
				break
			} else {
				fmt.Println("无效的输入，请重新输入。")
			}
		}

		// 如果设置为true，只有回答是"y"时才继续显示后面的问题
		if question.StopOnNo && strings.ToLower(question.Answer) != "y" {
			fmt.Println("配置完成: 选择无 API Key 模式")
			return
		}

		// 如果是最后一个问题，输出一个空行
		if i == len(s.Questions)-1 {
			fmt.Println()
		}
	}
}

func IsEmailValid(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func IsYesNoValid(answer string) bool {
	return answer == "y" || answer == "n"
}

// IsAPIKeyValid 验证 NCBI API key 是否符合格式要求 (包含字母、数字、破折号和下划线，并且长度在15到100之间)
func IsAPIKeyValid(apiKey string) bool {
	apiKeyRegex := `^[a-zA-Z0-9_-]{15,100}$`
	re := regexp.MustCompile(apiKeyRegex)
	return re.MatchString(apiKey)
}
