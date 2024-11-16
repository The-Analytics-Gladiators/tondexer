package models

import (
	"fmt"
)

type Dex string

const (
	stonfi = "stonfi"
	dedust = "dedust"
	all    = "all"
)

type DexParams struct {
	dexes []string
}

func ParseDex(s string) (Dex, error) {
	switch s {
	case string(stonfi):
		return stonfi, nil
	case string(dedust):
		return dedust, nil
	case string(all):
		return all, nil
	default:
		return all, nil
	}
}

func (dex Dex) WhereStatement(field string) string {
	switch dex {
	case stonfi:
		return fmt.Sprint("(", field, " = '", StonfiV1, "'",
			" OR ", field, " = '", StonfiV2, "'",
			")")
	case dedust:
		return fmt.Sprint("(", field, " = '", DeDust, "'", ")")
	case all:
		return fmt.Sprint("(", field, " = '", StonfiV1, "'",
			" OR ", field, " = '", StonfiV2, "'",
			" OR ", field, " = '", DeDust, "'",
			")")
	default:
		return ""
	}
}
