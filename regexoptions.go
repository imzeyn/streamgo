package streamgo

import (
	"regexp"
	"strings"
)

type RegexOptions struct {
	paramPatternRequired *regexp.Regexp
	paramPatternOptional *regexp.Regexp
	ParallelSearchCount  int
}

func NewRegexOptions(parallelSearchCount int) RegexOptions {
	return RegexOptions{
		paramPatternRequired: regexp.MustCompile(`^(.*?):[^/]+:`),
		paramPatternOptional: regexp.MustCompile(`^(.*?)::[^/]+::`),
		ParallelSearchCount:  parallelSearchCount,
	}
}

func (ro *RegexOptions) IsRequiredParam(name string) bool {
	return ro.paramPatternRequired.MatchString(name)
}

func (ro *RegexOptions) IsOptionalParam(name string) bool {
	return ro.paramPatternOptional.MatchString(name)
}

func (ro *RegexOptions) IsParamURL(name string) bool {
	return ro.IsRequiredParam(name) || ro.IsOptionalParam(name)
}

func (ro *RegexOptions) ParseParamNames(name string) map[int]string {
	mappedParams := make(map[int]string, 0)
	for i, v := range strings.Split(name, "/") {
		if !ro.IsParamURL(v) {
			continue
		}

		var paramName string
		if ro.IsOptionalParam(v) {
			paramName = ro.paramPatternOptional.FindStringSubmatch(v)[0]
		} else {
			paramName = ro.paramPatternRequired.FindStringSubmatch(v)[0]
		}

		mappedParams[i] = strings.ReplaceAll(paramName, ":", "")
		mappedParams[i] = strings.ReplaceAll(mappedParams[i], ":", "")
	}

	return mappedParams
}

func (ro *RegexOptions) ReplaceForFind(name string) string {
	for _, v := range ro.paramPatternOptional.FindAllString(name, -1) {
		name = strings.ReplaceAll(name, v, `?([\p{L}\p{N}\p{M}.@_-]*)?`)

	}

	for _, v := range ro.paramPatternRequired.FindAllString(name, -1) {
		name = strings.ReplaceAll(name, v, `[\p{L}\p{N}\p{M}.@_-]+`)
	}

	return name
}

func (ro *RegexOptions) GetPerfix(name string) string {
	var s strings.Builder

	for _, v := range name {
		if string(v) == ":" {
			break
		}
		s.WriteRune(v)
	}

	return s.String()
}
