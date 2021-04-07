package nevent

import (
	"fmt"
	"strings"
)

type SubjectTransformer func(subject string) string
type SubjectNormalizer func(subject string) string

func DefaultSubjectTransformer() SubjectTransformer {
	return func(subject string) string {
		return strings.ToLower(subject)
	}
}

func PatternSubjectTransformer(pattern string) SubjectTransformer {
	base := DefaultSubjectTransformer()
	return func(subject string) string {
		return fmt.Sprintf(pattern, base(subject))
	}
}

func STSubordinate(ns string) SubjectTransformer {
	return PatternSubjectTransformer("%s."+ns)
}

func STAny() SubjectTransformer {
	return STSubordinate("*")
}

func STGlob() SubjectTransformer {
	return STSubordinate(">")
}

func STValue(value string) SubjectTransformer {
	return STSubordinate(value)
}

func DefaultSubjectNormalize(subject string) string {
	s := subject
	s = strings.Replace(s, ".", "/", -1)
	s = strings.Replace(s, "*", "%", -1)
	s = strings.Replace(s, ">", "%%", -1)
	return s
}

func SubjectDenormalize(name string) string {
	s := name
	s = strings.Replace(s, "%%", ">", -1)
	s = strings.Replace(s, "%", "*", -1)
	s = strings.Replace(s, "/", ".", -1)
	return s
}

