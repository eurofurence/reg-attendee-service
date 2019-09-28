package config

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

func TestCheckConstraints(t *testing.T) {
	c := make(map[string]choiceConfig)
	c["selfref"] = choiceConfig{Constraint: "selfref", ConstraintMsg: "self referential"}
	c["msgmissing"] = choiceConfig{Constraint: "selfref"}
	c["wrongref"] = choiceConfig{Constraint: "unicorn", ConstraintMsg: "wrong reference"}

	actualErrors := url.Values{}
	checkConstraints(actualErrors, c, "blah", "selfref", c["selfref"].Constraint, c["selfref"].ConstraintMsg)
	checkConstraints(actualErrors, c, "blah", "msgmissing", c["msgmissing"].Constraint, c["msgmissing"].ConstraintMsg)
	checkConstraints(actualErrors, c, "blah", "wrongref", c["wrongref"].Constraint, c["wrongref"].ConstraintMsg)

	expectedErrors := url.Values{
		"blah.selfref.constraint":        []string{"invalid self referential constraint"},
		"blah.msgmissing.constraint_msg": []string{"blah.msgmissing.constraint_msg field must be at least 1 and at most 256 characters long"},
		"blah.wrongref.constraint":       []string{"invalid key in constraint, references nonexistent entry"},
	}
	prettyprintedActualErrors, _ := json.MarshalIndent(actualErrors, "", "  ")
	prettyprintedExpectedErrors, _ := json.MarshalIndent(expectedErrors, "", "  ")
	if !reflect.DeepEqual(actualErrors, expectedErrors) {
		t.Errorf("Errors were not as expected.\nActual:\n%v\nExpected:\n%v\n", string(prettyprintedActualErrors), string(prettyprintedExpectedErrors))
	}
}
