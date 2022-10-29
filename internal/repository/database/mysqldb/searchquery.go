package mysqldb

import (
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"sort"
	"strings"
)

func constructAttendeeSearchQuery(conds *attendee.AttendeeSearchCriteria, params map[string]interface{}) string {
	query := strings.Builder{}
	query.WriteString("SELECT * FROM attendees a WHERE (\n  (0 = 1)\n")
	if conds != nil {
		for i, cond := range conds.MatchAny {
			query.WriteString("  OR\n  (\n" + addSingleCondition(&cond, params, i+1) + "  )\n")
		}
	}
	query.WriteString(") ")
	if conds != nil {
		if conds.MinId > 0 {
			query.WriteString("AND a.id >= @param_0_1 ")
			params["param_0_1"] = conds.MinId
		}
		if conds.MaxId > 0 {
			query.WriteString("AND a.id <= @param_0_2 ")
			params["param_0_2"] = conds.MaxId
		}
		query.WriteString(orderBy(conds.SortBy, conds.SortOrder))
	}
	return query.String()
}

func orderBy(field string, direction string) string {
	if direction == "descending" {
		direction = "DESC"
	} else {
		direction = ""
	}
	var fieldName string
	switch field {
	case "birthday":
		fieldName = "a.birthday"
	case "city":
		fieldName = "a.city"
	case "country":
		fieldName = "a.country"
	case "email":
		fieldName = "a.email"
	case "name":
		fieldName = "CONCAT(a.first_name, ' ', a.last_name)"
	case "nickname":
		fieldName = "a.nickname"
	case "zip":
		fieldName = "a.zip"
	default:
		// status sort must be done in post
		fieldName = "a.id"
	}
	return fmt.Sprintf("ORDER BY %s %s", fieldName, direction)
}

func addSingleCondition(cond *attendee.AttendeeSearchSingleCriterion, params map[string]interface{}, idx int) string {
	paramBaseName := fmt.Sprintf("param_%d", idx)
	paramNo := 1
	query := strings.Builder{}
	query.WriteString("    (1 = 1)\n")
	if len(cond.Ids) > 0 {
		query.WriteString(uintSliceMatch("a.id", cond.Ids))
	}
	if cond.Nickname != "" {
		query.WriteString(fullstringMatch("a.nickname", cond.Nickname, params, paramBaseName, &paramNo))
	}
	if cond.Name != "" {
		query.WriteString(fullstringMatch("CONCAT(a.first_name, ' ', a.last_name)", cond.Name, params, paramBaseName, &paramNo))
	}
	if cond.Address != "" {
		query.WriteString(substringMatch("CONCAT(a.street, ' ', a.zip, ' ', a.city, ' ', a.state)", cond.Address, params, paramBaseName, &paramNo))
	}
	if cond.Country != "" {
		query.WriteString(stringExact("a.country", cond.Country, params, paramBaseName, &paramNo))
	}
	if cond.CountryBadge != "" {
		query.WriteString(stringExact("a.country_badge", cond.CountryBadge, params, paramBaseName, &paramNo))
	}
	if cond.Email != "" {
		query.WriteString(substringMatch("a.email", cond.Email, params, paramBaseName, &paramNo))
	}
	if cond.Telegram != "" {
		query.WriteString(substringMatch("a.telegram", cond.Telegram, params, paramBaseName, &paramNo))
	}
	query.WriteString(choiceMatch("a.flags", cond.Flags, params, paramBaseName, &paramNo))
	query.WriteString(choiceMatch("a.options", cond.Options, params, paramBaseName, &paramNo))
	query.WriteString(choiceMatch("a.packages", cond.Packages, params, paramBaseName, &paramNo))
	if cond.UserComments != "" {
		query.WriteString(substringMatch("a.user_comments", cond.UserComments, params, paramBaseName, &paramNo))
	}
	return query.String()
}

func uintSliceMatch(field string, values []uint) string {
	mappedValues := make([]string, len(values))
	for i, v := range values {
		mappedValues[i] = fmt.Sprintf("%d", v)
	}
	return fmt.Sprintf("    AND ( %s IN (%s))\n", field, strings.Join(mappedValues, ","))
}

func substringMatch(field string, condition string, params map[string]interface{}, paramBaseName string, idx *int) string {
	return fullstringMatch(field, "*"+condition+"*", params, paramBaseName, idx)
}

func fullstringMatch(field string, condition string, params map[string]interface{}, paramBaseName string, idx *int) string {
	mappedCondition := strings.ReplaceAll(condition, "*", "%")
	pName := fmt.Sprintf("%s_%d", paramBaseName, *idx)
	params[pName] = mappedCondition
	*idx++
	return fmt.Sprintf("    AND ( LOWER(%s) LIKE LOWER( @%s ) )\n", field, pName)
}

func stringExact(field string, condition string, params map[string]interface{}, paramBaseName string, idx *int) string {
	pName := fmt.Sprintf("%s_%d", paramBaseName, *idx)
	params[pName] = condition
	*idx++
	return fmt.Sprintf("    AND ( LOWER(%s) = LOWER( @%s ) )\n", field, pName)
}

func choiceMatch(field string, condition map[string]int8, params map[string]interface{}, paramBaseName string, idx *int) string {
	query := strings.Builder{}
	keys := make([]string, len(condition))
	i := 0
	for k := range condition {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, k := range keys {
		pName := fmt.Sprintf("%s_%d", paramBaseName, *idx)
		params[pName] = "%," + k + ",%"
		if condition[k] == 1 {
			query.WriteString(fmt.Sprintf("    AND ( %s LIKE @%s )\n", field, pName))
		} else if condition[k] == 0 {
			query.WriteString(fmt.Sprintf("    AND ( %s NOT LIKE @%s )\n", field, pName))
		}
		*idx++
	}
	return query.String()
}
