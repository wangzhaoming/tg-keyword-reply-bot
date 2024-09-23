package main

import (
	"fmt"
	"github.com/jpoz/groq"
	"regexp"
	"strconv"
	"strings"
	"tg-keyword-reply-bot/common"
	"tg-keyword-reply-bot/db"
)

const addText = "格式要求:\r\n" +
	"`/add 关键字===回复内容`\r\n\r\n" +
	"例如:\r\n" +
	"`/add 机场===https://jiji.cool`\r\n" +
	"就会添加一条规则, 关键词是机场, 回复内容是网址"
const delText = "格式要求:\r\n" +
	"`/del 关键字`\r\n\r\n" +
	"例如:\r\n" +
	"`/del 机场`\r\n" +
	"就会删除一条规则,机器人不再回复机场关键词"

/**
 * 添加规则
 */
func addRule(gid int64, rule string) {
	rules := common.AllGroupRules[gid]
	r := strings.Split(rule, "===")
	keys, value := r[0], r[1]
	if strings.Contains(keys, "||") {
		ks := strings.Split(keys, "||")
		for _, key := range ks {
			_addOneRule(key, value, rules)
		}
	} else {
		_addOneRule(keys, value, rules)
	}
	db.UpdateGroupRule(gid, rules.String())
}

/**
 * 给addRule使用的辅助方法
 */
func _addOneRule(key string, value string, rules common.RuleMap) {
	key = strings.Replace(key, " ", "", -1)
	rules[key] = value
}

/**
 * 删除规则
 */
func delRule(gid int64, key string) {
	rules := common.AllGroupRules[gid]
	delete(rules, key)
	db.UpdateGroupRule(gid, rules.String())
}

/**
 * 获取一个群组所有规则的列表
 */
func getRuleList(gid int64) []string {
	kvs := common.AllGroupRules[gid]
	str := ""
	var strs []string
	num := 1
	group := 0
	for k, v := range kvs {
		str += "\r\n\r\n规则" + strconv.Itoa(num) + ":\r\n`" + k + "` => `" + v + "` "
		num++
		group++
		if group == 10 {
			group = 0
			strs = append(strs, str)
			str = ""
		}
	}
	strs = append(strs, str)
	return strs
}

func getRuleListKey(gid int64) []string {
	kvs := common.AllGroupRules[gid]
	str := ""
	var strs []string
	for k, _ := range kvs {
		str += "`" + k + "`, "
	}
	strs = append(strs, str)
	return strs
}

/**
 * 查询是否包含相应的自动回复规则
 */
func findKey(gid int64, input string) string {
	kvs := common.AllGroupRules[gid]
	relate := ""
	if strings.HasPrefix(input, "~") || strings.HasPrefix(input, "～") {
		relate = relateKey(strings.TrimLeft(input, "~～"), getKeys(kvs))
	}
	for keyword, reply := range kvs {
		if strings.HasPrefix(keyword, "re:") {
			keyword = keyword[3:]
			match, _ := regexp.MatchString(keyword, input)
			if match {
				return reply
			}
		} else if input == keyword {
			return reply
		} else if relate == keyword {
			return reply
		}
	}
	return ""
}

func getKeys(m map[string]string) string {
	keys := ""
	for k := range m {
		keys += k + ","
	}
	return strings.Trim(keys, ",")
}

func relateKey(input string, target string) string {
	client := groq.NewClient(groq.WithAPIKey(apiKey))

	response, err := client.CreateChatCompletion(groq.CompletionCreateParams{
		Model: "llama-3.1-70b-versatile",
		Messages: []groq.Message{
			{
				Role: "user",
				Content: fmt.Sprintf(
					`
你是一个经验丰富的语言学家，精通各种语言知识，并且对网络流行词汇也有深入的研究。
现在你需要根据一个目标词，从多个候选词中，寻找一个和目标词最相关的词语，称之为关联词。
请考虑以下匹配逻辑以找到关联词语：
- 关联词可能和目标词语具有相似的功能或用途。
- 关联词可能和目标词语描述同一种事物或概念。
- 关联词可能和目标词语在特定上下文中存在关联性。
请确保考虑这些因素，并且只找到唯一一个最相关的关联词。

*注意*：只有候选词列表中的词语是有效的答案，回答时必须从候选词列表中进行选择，不要返回候选词之外的内容！不要返回目标词！特别的，找不到合适的答案时，允许返回null作为找不到的标识！

你在回答时，请使用方括号将找到的关联词包裹起来，并且只需要回答找到的词语即可，不需要添加多余的描述。

下面是一些示例：
示例1：
候选词[appleid,bingai,chatbot]，目标词"账户"
正确回答：[appleid]
示例2：
候选词[邮箱,bingai,chatbot]，目标词"抄送"
正确回答：[邮箱]
示例3：
候选词[appleid,中文,头像]，目标词"照片"
正确回答：[头像]
示例4：
候选词[appleid,中文,头像]，目标词"飞机"
正确回答：[null]
示例5：
候选词[appleid,中文,头像]，目标词"你好"
正确回答：[null]
示例6：
候选词[appleid,中文,头像]，目标词"饮料"
正确回答：[null]

下面是一些错误示例：
错误示例1：
候选词[appleid,中文,头像]，目标词"饮料"
错误回答：[饮料]
错误原因：饮料不在候选词中，不能返回候选词和null以外的内容。
错误示例2：
候选词[appleid,中文,头像]，目标词"你好"
错误回答：[hello]
错误原因：hello不在候选词中，不能返回候选词和null以外的内容。

现在，请从候选词[%s]中，找到与目标词语“%s”唯一相关的关联词。
`, target, input),
			},
		},
	})
	if err != nil {
		panic(err)
	}

	var reply = response.Choices[0].Message.Content
	reply = reply[1 : len(reply)-1]
	if reply == "null" {
		return ""
	}
	return reply
}
