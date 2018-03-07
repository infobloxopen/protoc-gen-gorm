package main

import (
	"fmt"
	"testing"
)

func TestLintName(t *testing.T) {
	for in, out := range map[string]string{
		"myId":        "myID",
		"VeryNormal":  "VeryNormal",
		"lowercase":   "lowercase",
		"go_uuid":     "goUUID",
		"myuuid":      "myuuid",
		"_":           "_",
		"word___id":   "wordID",
		"int___8___3": "int8_3",
		"uuid_id":     "uuidID",
	} {
		if linted := lintName(in); linted != out {
			t.Log(fmt.Sprintf("%s should have become %s, not %s", in, out, linted))
			t.Fail()
		}
	}
}
