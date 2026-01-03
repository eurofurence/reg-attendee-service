package config

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

func TestCheckConstraints(t *testing.T) {
	c := make(map[string]ChoiceConfig)
	c["selfref"] = ChoiceConfig{Constraint: "selfref", ConstraintMsg: "self referential"}
	c["msgmissing"] = ChoiceConfig{Constraint: "selfref"}
	c["wrongref"] = ChoiceConfig{Constraint: "unicorn", ConstraintMsg: "wrong reference"}
	c["admincrossref"] = ChoiceConfig{AdminOnly: true, Constraint: "!wrongref", ConstraintMsg: "wrong reference"}

	actualErrors := url.Values{}
	checkConstraints(actualErrors, c, "blah", "selfref", c["selfref"].Constraint, c["selfref"].ConstraintMsg)
	checkConstraints(actualErrors, c, "blah", "msgmissing", c["msgmissing"].Constraint, c["msgmissing"].ConstraintMsg)
	checkConstraints(actualErrors, c, "blah", "wrongref", c["wrongref"].Constraint, c["wrongref"].ConstraintMsg)
	checkConstraints(actualErrors, c, "blah", "admincrossref", c["admincrossref"].Constraint, c["admincrossref"].ConstraintMsg)

	expectedErrors := url.Values{
		"blah.admincrossref.constraint":  []string{"invalid key in constraint, references across admin only and non-admin only"},
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

func TestCheckFlags(t *testing.T) {
	c := make(map[string]ChoiceConfig)
	c["admindefault"] = ChoiceConfig{Default: true, AdminOnly: true, Description: "admin and default at the same time - invalid"}
	c["adminro"] = ChoiceConfig{AdminOnly: true, ReadOnly: true, Description: "admin and read only at the same time - invalid"}

	actualErrors := url.Values{}
	validateFlagsConfiguration(actualErrors, c)
	expectedErrors := url.Values{
		"choices.flags.admindefault.default": []string{"a flag cannot both be admin_only and default to on"},
		"choices.flags.adminro.admin":        []string{"a flag cannot both be admin_only and read_only"},
	}
	prettyprintedActualErrors, _ := json.MarshalIndent(actualErrors, "", "  ")
	prettyprintedExpectedErrors, _ := json.MarshalIndent(expectedErrors, "", "  ")
	if !reflect.DeepEqual(actualErrors, expectedErrors) {
		t.Errorf("Errors were not as expected.\nActual:\n%v\nExpected:\n%v\n", string(prettyprintedActualErrors), string(prettyprintedExpectedErrors))
	}
}

func TestCheckPackages(t *testing.T) {
	c := make(map[string]ChoiceConfig)
	c["myadmin"] = ChoiceConfig{Default: true, AdminOnly: true, Description: "admin only package - invalid"}
	c["counttoohigh"] = ChoiceConfig{AllowedCounts: []int{1, 17, 34}, Description: "allowed_counts higher than max_count", MaxCount: 17}
	c["maxcountunset"] = ChoiceConfig{AllowedCounts: []int{1, 17, 34}, Description: "allowed_counts but no max_count"}
	c["limitnegative"] = ChoiceConfig{Description: "limit negative", Limit: -4}

	actualErrors := url.Values{}
	validatePackagesConfiguration(actualErrors, c)
	expectedErrors := url.Values{
		"choices.packages.myadmin.admin":                []string{"packages cannot be admin_only (they cost money). Try read_only instead."},
		"choices.packages.counttoohigh.allowed_counts":  []string{"maximum allowed_counts value cannot be larger than max_count for package"},
		"choices.packages.maxcountunset.allowed_counts": []string{"can only list allowed counts if max_count is set to at least 2"},
		"choices.packages.limitnegative.limit":          []string{"limit value cannot be negative"},
	}
	prettyprintedActualErrors, _ := json.MarshalIndent(actualErrors, "", "  ")
	prettyprintedExpectedErrors, _ := json.MarshalIndent(expectedErrors, "", "  ")
	if !reflect.DeepEqual(actualErrors, expectedErrors) {
		t.Errorf("Errors were not as expected.\nActual:\n%v\nExpected:\n%v\n", string(prettyprintedActualErrors), string(prettyprintedExpectedErrors))
	}
}

func TestCheckOptions(t *testing.T) {
	c := make(map[string]ChoiceConfig)
	c["myadmin"] = ChoiceConfig{Default: true, AdminOnly: true, Description: "admin only option - invalid"}
	c["myro"] = ChoiceConfig{ReadOnly: true, Description: "read only option - invalid"}

	actualErrors := url.Values{}
	validateOptionsConfiguration(actualErrors, c)
	expectedErrors := url.Values{
		"choices.options.myadmin.admin": []string{"options cannot be admin_only (they represent user choices)."},
		"choices.options.myro.readonly": []string{"options cannot be read_only (they represent user choices)."},
	}
	prettyprintedActualErrors, _ := json.MarshalIndent(actualErrors, "", "  ")
	prettyprintedExpectedErrors, _ := json.MarshalIndent(expectedErrors, "", "  ")
	if !reflect.DeepEqual(actualErrors, expectedErrors) {
		t.Errorf("Errors were not as expected.\nActual:\n%v\nExpected:\n%v\n", string(prettyprintedActualErrors), string(prettyprintedExpectedErrors))
	}
}
