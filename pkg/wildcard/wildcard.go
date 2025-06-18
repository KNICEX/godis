package wildcard

import (
	"errors"
	"regexp"
	"strings"
)

var replaceMap = map[byte]string{
	'+': `\+`,
	')': `\)`,
	'$': `\$`,
	'.': `\.`,
	'{': `\{`,
	'}': `\}`,
	'|': `\|`,
	'*': `.*`,
	'?': `.`,
}

type Pattern struct {
	exp *regexp.Regexp
}

func (p *Pattern) Match(s string) bool {
	return p.exp.MatchString(s)
}

func Compile(src string) (*Pattern, error) {
	regexSrc := strings.Builder{}
	_ = regexSrc.WriteByte('^')
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if ch == '\\' {
			if i == len(src)-1 {
				return nil, errors.New("end with escape character")
			}
			_ = regexSrc.WriteByte(ch)
			_ = regexSrc.WriteByte(src[i+1])
			i++
		} else if ch == '^' {
			if i == 0 {
				_, _ = regexSrc.WriteString(`\^`)
			} else if i == 1 {
				if src[i-1] == '[' {
					_, _ = regexSrc.WriteString(`^`)
				} else {
					_, _ = regexSrc.WriteString(`\^`)
				}
			} else {
				if src[i-1] == '[' && src[i-2] != '\\' {
					_, _ = regexSrc.WriteString(`^`)
				} else {
					_, _ = regexSrc.WriteString(`\^`)
				}
			}
		} else if escaped, toEscape := replaceMap[ch]; toEscape {
			_, _ = regexSrc.WriteString(escaped)
		} else {
			_ = regexSrc.WriteByte(ch)
		}
	}

	_ = regexSrc.WriteByte('$')
	exp, err := regexp.Compile(regexSrc.String())
	if err != nil {
		return nil, err
	}
	return &Pattern{
		exp: exp,
	}, nil
}
