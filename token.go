package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Drelf2020/utils"
)

// 单条语句
//
// 其中 Args 表示其 Name 中携带的参数
//
// 当该语句为字典时 Tokens 不为空 表示其下包含的语句
type Token struct {
	Type   string            `json:"type,omitempty" yaml:"type,omitempty"`
	Name   string            `json:"-" yaml:"-"`
	Hint   string            `json:"hint,omitempty" yaml:"hint,omitempty"`
	Value  string            `json:"-" yaml:"-"`
	Output any               `json:"value,omitempty" yaml:"value,omitempty"`
	Parent *Token            `json:"-" yaml:"-"`
	Args   []string          `json:"-" yaml:"-"`
	Params []string          `json:"-" yaml:"-"`
	Items  []*Token          `json:"-" yaml:"-"`
	Tokens map[string]*Token `json:"-" yaml:"-"`
}

// 判断该语句是否为 Api 起始语句
func (token *Token) IsApi() bool {
	return MethodTypes.Has(token.Type)
}

// 判断该语句是否为 type 起始语句
func (token *Token) IsType() bool {
	return token.Type == "type"
}

// 判断该语句是否为枚举
func (token *Token) IsEnum() bool {
	return token.Type == "enum"
}

// 判断该语句是否为起始括号
func (token *Token) IsOpen() bool {
	return token.Value == "{"
}

// 判断该语句是否为起始列表
func (token *Token) IsBracket() bool {
	return token.Value == "["
}

// 判断该语句是否为闭合括号
func (token *Token) IsClose() bool {
	return token.Name == "}" || token.Name == "]"
}

// 判断该语句是否为字典
func (token *Token) IsDict() bool {
	return token.IsOpen() || len(token.Tokens) != 0 || token.Type == "dict"
}

// 判断该语句是否为列表
func (token *Token) IsList() bool {
	return token.IsBracket() || len(token.Items) != 0 || token.Type == "list" || utils.Startswith(token.Type, "[")
}

// 判断是否为多行文本
func (token *Token) IsMultiLine() string {
	if len(token.Value) < 1 {
		return ""
	}
	if chn := token.Value[:1]; In([]string{"\"", "'", "`"}, chn) {
		return chn
	}
	return ""
}

// 判断是否有引号
func (token *Token) HasQuotation(chn string) bool {
	return strings.Count(token.Name, chn) == 1
}

// 判断该语句是否为必要变量
func (token *Token) IsRequired() bool {
	return token.Value == ""
}

// 判断该语句是否为常量
func (token *Token) IsConstant() bool {
	return utils.Endswith(token.Value, ",constant")
}

// 判断该语句是否为选填变量
func (token *Token) IsOptional() bool {
	return (token.Value == "none") || !token.IsRequired() && !token.IsConstant()
}

// 添加子语句
func (token *Token) Add(t *Token, isList ...bool) {
	if token == nil {
		return
	}
	t.Parent = token
	if token.IsList() || (len(isList) != 0 && isList[0]) {
		token.Items = append(token.Items, t)
	} else {
		token.Tokens[t.Name] = t
	}
}

// 获取子语句
func (token *Token) Get(name string) *Token {
	return token.Tokens[name]
}

// 移除子语句
func (token *Token) Pop(name string) *Token {
	if v := token.Get(name); v != nil {
		delete(token.Tokens, name)
		return v
	}
	return nil
}

// 交换类型和变量名
func (token *Token) Exchange(vt *Types) *Token {
	return token.Copy(token.Name+"<"+strings.Join(token.Args, ",")+">", vt)
}

// 转字典
func (token *Token) ToDict() map[string]string {
	dic := make(map[string]string)
	utils.ForMap(token.Tokens, func(s string, t *Token) { dic[s] = t.Value })
	return dic
}

// 复制
func (token *Token) Copy(nt string, vt *Types) *Token {
	t := NewToken(nt, token.Name, token.Hint, token.Value)
	t.SetTypes(vt)
	return t
}

// 修改输出
func (token *Token) SetOutput() {
	if token.IsList() {
		token.Output = &token.Items
	} else if token.IsDict() {
		token.Output = token.Tokens
	}
}

// 返回类型数组的具体内容和长度
func (token *Token) GetLength(value ...string) (string, int64) {
	var typ string = token.Type
	var length int64 = -1

	if len(value) != 0 {
		typ = value[0]
	}

	// 判断是否为数组
	utils.ForEach(
		regexp.MustCompile(`\[(\d*)\]`).FindAllStringSubmatch(typ, -1),
		func(s []string) {
			typ = strings.Replace(typ, s[0], "", 1)
			length, _ = strconv.ParseInt(s[1], 10, 64)
			if length < 1 {
				length = 1
			}
		},
		func(s []string) bool { return len(s) != 0 },
	)

	return typ, length
}

// 修改类型
func (token *Token) SetTypes(vt *Types) {
	typ, length := token.GetLength()
	tk := vt.Get(typ)

	// 基础类型 不做修改 输出原值
	if tk == nil {
		if length != -1 {
			tk = NewToken(typ, "", "", "")
			for i := 0; i < int(length); i++ {
				token.Add(tk, true)
			}
		}
		// 有子语句(dict | list) 替换输出为子语句集合
		token.SetOutput()
		return
	}

	// 枚举 替换输出为具体值
	if tk.IsEnum() {
		token.Output = tk.Get(token.Value).Value
		return
	}

	// 自定义类型
	var nt *Token
	argsMap := NewZip(tk.Args, token.Params, token.Type+" 的参数个数都能数歪来？")

	Copy := func(t *Token) *Token {
		p := ""
		if len(t.Params) != 0 {
			p = "<" + strings.Join(t.Params, ",") + ">"
		}
		return t.Copy(argsMap.Same(t.Type)+p, vt)
	}

	if length != -1 {
		nt = NewToken(tk.Name, tk.Name, tk.Hint, tk.Value)
	} else {
		nt = token
	}
	if tk.IsDict() {
		utils.ForMap(tk.Tokens, func(s string, t *Token) { nt.Add(Copy(t)) })
	} else if tk.IsList() {
		utils.ForEach(tk.Items, func(t *Token) { nt.Add(Copy(t), true) })
	}
	if length != -1 {
		nt.SetOutput()
		for i := 0; i < int(length); i++ {
			token.Add(nt, true)
		}
	}

	token.SetOutput()
}

// 解析类似 res<T1,T2> 中的子类型 T1 T2
//
// data 顺序 Sentence Type Name Hint Value
func NewToken(data ...string) *Token {
	// 自动推断类型
	typ, output := AutoType(data[0], data[3])
	// 解析类型中的参数
	typ, params := NameSlice(typ)
	// 解析变量名中的参数
	name, args := NameSlice(strings.TrimSpace(data[1]))
	// 去除标注前后空白
	hint := strings.TrimSpace(data[2])

	return &Token{typ, name, hint, data[3], output, nil, args, params, make([]*Token, 0), make(map[string]*Token)}
}
