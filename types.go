package parser

import (
	"regexp"
	"strings"

	"github.com/Drelf2020/utils"
)

// 类型集合
type Types map[string]*Token

// 添加类型
func (types Types) Add(token *Token, names ...string) {
	for _, n := range names {
		types[n] = nil
	}
	if token != nil {
		typ, length := token.GetLength(token.Value)
		if length != -1 {
			t := NewToken(typ, "", "", "")
			t.SetTypes(&types)
			for i := 0; i < int(length); i++ {
				token.Add(t, true)
			}
			token.Value = "List<" + typ + ">"
		}
		types[token.Name] = token
	}
}

// 判断字段
func (types *Types) Has(key string) bool {
	for k := range *types {
		if key == k {
			return true
		}
	}
	return false
}

// 以字符串形式连接多个 type
func (types *Types) Join(keys ...string) string {
	for k := range *types {
		keys = append(keys, k)
	}
	return strings.Join(keys, "|")
}

// 合并多个类型组
func (ts *Types) Union(typess ...*Types) *Types {
	nt := NewTypes()
	for k, v := range *ts {
		nt.Add(v, k)
	}
	for _, types := range typess {
		for k, v := range *types {
			nt.Add(v, k)
		}
	}
	return nt
}

// 生成正则表达式
func (types *Types) ToRegexp() *regexp.Regexp {
	return regexp.MustCompile(` *(?:((?:\[\d*\])?(?:` + types.Join() + `)<?(?:` + types.Join(",", "<", ">", `\[`, `\]`, `\w`) + `)*>?) )? *([^:^=^\r^\n]+)(?:: *([^=^\r^\n]+))? *(?:= *([^\r^\n]+))?`)
}

// 正则查找语句
func (types *Types) FindTokens(api string) (tokens []*Token) {
	re := types.ToRegexp()
	for _, sList := range re.FindAllStringSubmatch(api, -1) {
		if !utils.Startswith(strings.TrimSpace(sList[0]), "#") {
			tokens = append(tokens, NewToken(sList[1:]...))
		}
	}
	return
}

// 获取 Token
//
// 当 key 为基础类型(str num bool)时返回 nil
func (types *Types) Get(key string) *Token {
	return (*types)[key]
}

// 构造函数
func NewTypes(keys ...string) *Types {
	types := make(Types)
	types.Add(nil, keys...)
	return &types
}

// 支持的请求类型 GET POST
var MethodTypes = NewTypes("GET", "POST")
