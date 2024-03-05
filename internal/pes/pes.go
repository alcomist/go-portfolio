// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pes

import (
	"bytes"
	"fmt"
	"github.com/alcomist/go-portfolio/internal/stack"
	"github.com/alcomist/go-portfolio/internal/util"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// 스크립트 작성요령

// DF : 기본값 설정
// 명령어 예시
// DF:unit=>1
// DF:box=>1
// DF:kilo=>9999
// DF:kilo=>9999,unit=>100

// RM : 해당 단어 혹은 패턴으로 주어진 문자열에서 삭제
// 명령어 예시
// RM:단어
// RM:단어1,단어2,단어3
// RM:Dp+Dp

// RP : 해당 단어 혹은 패턴으로 주어진 문자열 치환
// 명령어 예시
// RP:Dunit=>&1갯수
// RP:Dbox=>&1박스

// ET : 해당 패턴으로 주어진 문자열을 환경변수 map에 (env)에 값 추가
// 명령어 예시
// ET:Dunit=>&1=unit
// ET:Dbox=>&1=box
// ET:Dkilo=>&1=kilo

// ST:brand=>브랜드=브랜드, SET:brand=>브랜드=, SET:brand=>브랜드=brand

// LM : 지정된 키의 값을 제한한다.
// 명령어 예시
// LM:unit=>gt1,lte100",
// LM:box=>gt1,lte10"

// DF, RM, RP, ET, LM 의 순서로 작성
// DF는 전체 명령어 첫부분에
// LM은 전체 명령어 맨 마지막에 일괄적으로 실행

// ET 명령은 RM, RP 명령어 다음에 처리해야 정상적으로 추출이 가능

type Env map[string]string

type EnvSet struct {
	Start int
	End   int
	Key   string
	Value string
	Alt   string
}

func (es *EnvSet) String() string {

	var b bytes.Buffer

	fmt.Fprintf(&b, "START=%v\n", es.Start)
	fmt.Fprintf(&b, "END=%v\n", es.End)
	fmt.Fprintf(&b, "KEY=%v\n", es.Key)
	fmt.Fprintf(&b, "VAL=%v\n", es.Value)
	fmt.Fprintf(&b, "ALT=%v\n", es.Alt)

	return b.String()
}

type Envs []*EnvSet

func (x Envs) Len() int           { return len(x) }
func (x Envs) Less(i, j int) bool { return x[i].Start < x[j].Start }
func (x Envs) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

const (
	OpcodeDefault = "df"
	OpcodeRemove  = "rm"
	OpcodeReplace = "rp"
	OpcodeExtract = "et"
	OpcodeSet     = "st"
	OpcodeLimit   = "lm"
)

type instruction struct {
	opcode   string
	arg      string
	operands [][2]string
}

func (i *instruction) String() string {

	var b bytes.Buffer

	fmt.Fprintf(&b, "OPCODE=%v\n", i.opcode)
	for _, op := range i.operands {
		fmt.Fprintf(&b, "\tSOURCE=%v\n", op[0])
		fmt.Fprintf(&b, "\tDESTINATION=%v\n", op[1])
	}

	return b.String()
}

type Callstack struct {
	Opcode  string `json:"opcode"`
	Arg     string `json:"args"`
	Pre     string `json:"pre"`
	Pro     string `json:"pro"`
	Changed bool   `json:"changed"`
}

func (s *Callstack) String() string {

	var b bytes.Buffer

	fmt.Fprintf(&b, "OPCODE=%v\n", s.Opcode)
	fmt.Fprintf(&b, "ARG=%v\n", s.Arg)
	fmt.Fprintf(&b, "PRE=%v\n", s.Pre)
	fmt.Fprintf(&b, "PRO=%v\n", s.Pro)
	fmt.Fprintf(&b, "CHANGED=%v\n", s.Changed)

	return b.String()
}

func isValidOpcode(c string) bool {

	c = strings.ToLower(c)

	codes := []string{
		OpcodeDefault,
		OpcodeRemove,
		OpcodeReplace,
		OpcodeExtract,
		OpcodeSet,
		OpcodeLimit,
	}

	return util.Contains(c, codes)
}

func decode(input string) *instruction {

	input = strings.Trim(input, " ")

	if strings.HasPrefix(input, "#") || strings.HasPrefix(input, "//") {
		return nil
	}

	parts := strings.Split(input, ":")
	if len(parts) < 2 {
		return nil
	}

	var inst instruction

	opcode := strings.ToLower(parts[0])

	if isValidOpcode(opcode) == false {
		return nil
	}

	// command must be in lower case
	inst.opcode = opcode
	inst.arg = parts[1]
	inst.operands = make([][2]string, 0)

	for _, arg := range strings.Split(parts[1], ",") {
		args := strings.Split(arg, "=>")

		op := [2]string{}

		if len(args) == 1 {
			op[0] = strings.Trim(args[0], " ")
		} else if len(args) == 2 {
			op[0], op[1] =
				strings.Trim(args[0], " "), strings.Trim(args[1], " ")
		}

		inst.operands = append(inst.operands, op)
	}

	return &inst
}

func metaKey(k string) string {
	return fmt.Sprintf("meta_%s", k)
}

func calculateExpression(s string) string {

	// 정수형 연산 패턴은 아래와 같음
	//p := `(\d+(\s*[*+\-/]\s*\d+)+)`

	// 실수형 연산을 위해 정수형 패턴을 확장함
	p := `(\d+(\.\d+)?(\s*[*.+\-/]\s*\d+(\.\d+)?)+)`

	r, err := regexp.Compile(p)
	if err != nil {
		log.Println(err)
		return s
	}

	matches := r.FindAllString(s, -1)

	for _, match := range matches {

		m := match

		m = util.SpaceOperators(m)

		splits := strings.Fields(m)

		st := stack.NewFloatStack()
		op := ""

		for _, split := range splits {
			if util.Contains(split, []string{"+", "-", "/", "*"}) {
				op = split
			} else {
				val, err := strconv.ParseFloat(split, 64)
				if err != nil {
					log.Println(err)
				}

				if st.Len() > 0 {
					prev := st.Pop()

					switch op {
					case "*":
						st.Push(prev * val)
					case "-":
						st.Push(prev - val)
					case "+":
						st.Push(prev + val)
					case "/":
						if val != 0 {
							st.Push(prev / val)
						}
					}
				} else {
					st.Push(val)
				}
			}
		}

		// 소수점 아래 3자리까지 표시하도록 함
		// 소수점 아래가 0일 경우 정수형처럼 표시
		// 33.33333333333 => 33.333
		// 28 => 28
		// 0.32 => 0.32
		// 1.333 => 1.333
		// 1.423423 => 1.423

		if st.Len() == 1 {
			res := fmt.Sprintf("%g", st.Pop())
			pos := strings.Index(res, ".")
			if pos != -1 {
				if len(res) > pos+4 {
					res = res[:pos+4]
				}
			}

			s = strings.ReplaceAll(s, match, res)
		}
	}

	return s
}

func replaceRef(env Env, dst string) string {

	p := `(&\s*\d+)`
	r, err := regexp.Compile(p)
	if err != nil {
		log.Println(err)
		return dst
	}

	// find index pattern in the dst and replace with env value
	matches := r.FindAllString(dst, -1)

	for _, match := range matches {

		m := util.RemoveSpace(match)

		val, ok := env[m[1:]]
		if ok {
			dst = strings.ReplaceAll(dst, match, val)
		} else {
			dst = strings.ReplaceAll(dst, match, "0")
		}
	}

	// calculate expression in the dst and replace
	dst = calculateExpression(dst)

	return dst
}

func Set(s, src, dst string) (string, []*EnvSet) {

	sets := make([]*EnvSet, 0)

	ss := strings.Split(dst, "=")

	search := ss[0]
	target := ss[0]

	if len(ss) > 1 {
		target = ss[1]
	}

	p := util.PatternToRegex(search)

	r, err := regexp.Compile(p)
	if err != nil {
		fmt.Println(err)
		return s, sets
	}

	matches := r.FindAllStringIndex(s, -1)
	if len(matches) == 0 {
		return s, sets
	}

	for _, match := range matches {

		b, e := match[0], match[1]
		v := s[b:e]

		sets = append(sets, &EnvSet{Start: b, End: e, Key: src, Value: v, Alt: target})
	}

	for _, set := range sets {
		s = strings.ReplaceAll(s, set.Value, strings.Repeat(" ", utf8.RuneCountInString(set.Value)))
	}

	return s, sets
}

func Replace(s, src, dst string) string {

	p := util.PatternToRegex(src)

	r, err := regexp.Compile(p)
	if err != nil {
		fmt.Println(err)
		return s
	}

	matches := r.FindStringSubmatch(s)

	if len(matches) == 0 {
		return s
	}

	// for remove operation
	if len(dst) == 0 {
		return strings.Replace(s, matches[0], dst, -1)
	}

	// set env for referenced replace
	env := make(Env)
	for i, match := range matches[1:] {
		env[fmt.Sprintf("%d", i+1)] = match
	}

	dst = replaceRef(env, dst)

	return strings.ReplaceAll(s, matches[0], dst)
}

func extractRef(env Env, dst string) (string, Env) {

	p := `(&\s*\d+=\w+)`
	r, err := regexp.Compile(p)
	if err != nil {
		return dst, nil
	}

	// find index pattern in the dst and replace with env value
	matches := r.FindAllString(dst, -1)
	if len(matches) == 0 {
		return dst, nil
	}

	res := make(Env)

	for _, match := range matches {

		match = util.RemoveSpace(match)
		split := strings.Split(match, "=")

		if len(split) == 2 && len(split[1]) > 0 {
			val, ok := env[split[0][1:]]
			if ok {
				res[split[1]] = val
				dst = strings.ReplaceAll(dst, match, "")
			}
		}
	}

	return dst, res
}

func Extract(s, src, dst string) (string, Env) {

	p := util.PatternToRegex(src)

	r, err := regexp.Compile(p)
	if err != nil {
		fmt.Println(err)
		return s, nil
	}

	matches := r.FindStringSubmatch(s)
	if len(matches) == 0 {
		return s, nil
	}

	// for remove operation
	if len(dst) == 0 {
		return s, nil
	}

	// set env for referenced replace
	env := make(Env)
	for i, match := range matches {
		env[fmt.Sprintf("%d", i)] = match
	}

	// 2D와 같은 패턴을 쓰는 것이 아닌 순수 텍스트를 추출하려고 추가 예외 처리
	if len(matches) == 1 && len(matches[0]) > 0 && len(env) == 1 {
		env["1"] = matches[0]
	}

	dst, res := extractRef(env, dst)

	return strings.ReplaceAll(s, matches[0], dst), res
}

// gt '>'
// gte '>='
// lt '<'
// lte '<='

func applyConditions(f float64, dst string) (float64, bool) {

	conditions := strings.Split(dst, ",")
	if len(conditions) > 2 {
		return f, false
	}

	changed := false
	ops := []string{"gte", "lte", "gt", "lt"}

	for _, condition := range conditions {
		condition = util.SpaceNumeric(condition)
		cmp := strings.Fields(condition)
		if len(cmp) == 2 {
			op := cmp[0]
			if util.Contains(op, ops) {
				v, e := strconv.ParseFloat(cmp[1], 64)
				if e != nil {
					fmt.Println(e)
				}
				switch op {
				case "gte":
					if !(v <= f) {
						f = v
						changed = true
					}
				case "lte":
					if !(f <= v) {
						f = v
						changed = true
					}
				case "gt":
					if !(v < f) {
						f = v + 1
						changed = true
					}
				case "lt":
					if !(f < v) {
						f = v - 1
						changed = true
					}
				}
			}
		}
	}

	return f, changed
}

func Limit(env Env, src, dst string) Env {

	val, ok := env[src]
	if !ok {
		return env
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		fmt.Println(err)
		return env
	}

	f, changed := applyConditions(f, dst)
	if changed {
		env[src] = fmt.Sprintf("%g", f)
	}

	return env
}

func Default(env Env, op [2]string) {

	if len(op[0]) > 0 && len(op[1]) > 0 {
		env[op[0]] = op[1]
	}
}

func DeepCopy(e Env) Env {
	org := Env{}
	for k, v := range e {
		org[k] = v
	}

	return org
}

func DeepEqual(m1, m2 Env) bool {

	// m1 and m2 both are nil
	if m1 == nil && m2 == nil {
		return true
	}

	// m1 and m2 both are 0 length
	if len(m1) == 0 && len(m2) == 0 {
		return true
	}

	if m1 == nil || m2 == nil {
		return false
	}

	if len(m1) == 0 || len(m2) == 0 {
		return false
	}

	for k, v := range m1 {
		v2, ok := m2[k]
		if ok {
			if v != v2 {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// Sandbox implementation

type Sandbox struct {
	defaults []*instruction
	commands []*instruction
	limits   []*instruction

	debug      bool
	callstacks []*Callstack

	env  Env
	envs Envs
}

func NewSandbox() *Sandbox {

	s := Sandbox{}
	return &s
}

func (sb *Sandbox) Debug(debug bool) {
	sb.debug = debug
}

func (sb *Sandbox) Decode(inputs []string) {

	sb.commands = make([]*instruction, 0, len(inputs))
	sb.limits = make([]*instruction, 0)

	for _, input := range inputs {

		a := decode(input)
		if a == nil {
			continue
		}

		if a.opcode == OpcodeDefault {
			sb.defaults = append(sb.defaults, a)
		} else if a.opcode == OpcodeLimit {
			sb.limits = append(sb.limits, a)
		} else {
			sb.commands = append(sb.commands, a)
		}
	}
}

func (sb *Sandbox) Execute(ss []string) {

	sb.callstacks = make([]*Callstack, 0)
	sb.env = make(Env)
	sb.envs = make([]*EnvSet, 0)

	for _, s := range ss {
		sb.execute(s)
	}
}

func (sb *Sandbox) execute(s string) {

	for _, def := range sb.defaults {
		for _, op := range def.operands {
			Default(sb.env, op)
		}
	}

	for _, mid := range sb.commands {

		if len(s) == 0 {
			break
		}

		cs := Callstack{}
		cs.Opcode = mid.opcode
		cs.Arg = mid.arg

		for _, op := range mid.operands {

			switch mid.opcode {
			case OpcodeRemove:
				cs.Pre = s
				s = Replace(s, op[0], "")
				cs.Pro = s
				if cs.Pre != cs.Pro {
					cs.Changed = true
				}
			case OpcodeReplace:
				cs.Pre = s
				s = Replace(s, op[0], op[1])
				cs.Pro = s
				if cs.Pre != cs.Pro {
					cs.Changed = true
				}
			case OpcodeExtract:
				cs.Pre = s
				s0, r := Extract(s, op[0], op[1])
				cs.Pro = s0
				s = s0
				if cs.Pre != cs.Pro {
					cs.Changed = true
				}
				for k, v := range r {
					sb.env[k] = v
				}
			case OpcodeSet:
				cs.Pre = s
				s0, sets := Set(s, op[0], op[1])
				cs.Pro = s0
				s = s0
				if cs.Pre != cs.Pro {
					cs.Changed = true
				}

				sb.envs = append(sb.envs, sets...)
			}
		}

		if sb.debug {
			sb.callstacks = append(sb.callstacks, &cs)
		}
	}

	for _, lim := range sb.limits {
		for _, op := range lim.operands {
			Limit(sb.env, op[0], op[1])
		}
	}
}

func (sb *Sandbox) Meta() Env {

	e := Env{}
	for k, v := range sb.env {
		e[metaKey(k)] = v
	}

	return e
}

func (sb *Sandbox) Env() Env {

	e := Env{}
	for k, v := range sb.env {
		e[k] = v
	}

	// 복사본 전달
	return e
}

func (sb *Sandbox) Envs() Envs {

	sort.Sort(sb.envs)
	return sb.envs
}

func (sb *Sandbox) Callstacks() []*Callstack {
	return sb.callstacks
}
