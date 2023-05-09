package mysqldb

import (
	"context"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"sort"
	"strings"
)

func (r *MysqlRepository) constructAttendeeSearchQuery(ctx context.Context, conds *attendee.AttendeeSearchCriteria, params map[string]interface{}) string {
	newestStatusSubQuery := strings.Builder{}
	newestStatusSubQuery.WriteString(" SELECT sc.attendee_id AS attendee_id, ")
	newestStatusSubQuery.WriteString("        ( SELECT sc2.status FROM att_status_changes AS sc2 WHERE sc2.id = max(sc.id) ) AS status ")
	newestStatusSubQuery.WriteString(" FROM att_status_changes AS sc ")
	newestStatusSubQuery.WriteString(" GROUP BY sc.attendee_id ")

	query := strings.Builder{}
	query.WriteString("SELECT ")
	fields := constructFieldList(conds.FillFields)
	query.WriteString(strings.Join(fields, ", ") + " \n")
	query.WriteString("FROM att_attendees AS a \n")
	query.WriteString("  LEFT JOIN att_admin_infos AS ad ON ad.id = a.id \n")
	query.WriteString("  LEFT JOIN ( " + newestStatusSubQuery.String() + " ) AS st ON st.attendee_id = a.id \n")
	query.WriteString("WHERE (\n  (0 = @param_force_named_query_detection)\n")
	params["param_force_named_query_detection"] = 1 // if we do not have at least one @-param, GORM doesn't do named query mode, and then it fails with a type error
	if conds != nil {
		for i, cond := range conds.MatchAny {
			query.WriteString("  OR\n  (\n" + r.addSingleCondition(&cond, params, i+1) + "  )\n")
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
	result := query.String()
	aulogging.Logger.Ctx(ctx).Debug().Printf("SQL query: %s", result)
	return result
}

func constructFieldList(spec []string) []string {
	selected := make(map[string]bool)
	if len(spec) == 0 {
		// default selection
		selected["a.id as id"] = true
		selected["a.nickname as nickname"] = true
		selected["a.first_name as first_name"] = true
		selected["a.last_name as last_name"] = true
		selected["a.country as country"] = true
		selected["a.spoken_languages as spoken_languages"] = true
		selected["a.email as email"] = true
		selected["a.telegram as telegram"] = true
		selected["a.birthday as birthday"] = true
		selected["a.pronouns as pronouns"] = true
		selected["a.tshirt_size as tshirt_size"] = true
		selected["a.flags as flags"] = true
		selected["IFNULL(ad.flags, '') as admin_flags"] = true
		selected["a.options as options"] = true
		selected["a.packages as packages"] = true
		selected["a.user_comments as user_comments"] = true
		selected["IFNULL(st.status, 'new') as status"] = true
		selected["a.cache_total_dues as cache_total_dues"] = true
		selected["a.cache_payment_balance as cache_payment_balance"] = true
		selected["a.cache_open_balance as cache_open_balance"] = true
		selected["a.cache_due_date as cache_due_date"] = true
		selected["a.created_at as created_at"] = true
		selected["IFNULL(ad.admin_comments, '') as admin_comments"] = true
	} else {
		selected["a.id as id"] = true // always show badge number
		for _, s := range spec {
			defKey := fmt.Sprintf("a.%s as %s", s, s)
			switch s {
			case "id", "nickname", "first_name", "last_name", "street", "zip", "city":
				selected[defKey] = true
			case "country", "spoken_languages", "registration_language", "state", "email", "phone", "telegram", "partner":
				selected[defKey] = true
			case "birthday", "gender", "pronouns", "tshirt_size":
				selected[defKey] = true
			case "flags":
				selected["a.flags as flags"] = true
				selected["IFNULL(ad.flags, '') as admin_flags"] = true // from search viewpoint, these are just more flags
			case "options", "packages", "user_comments":
				selected[defKey] = true
			case "status":
				selected["IFNULL(st.status, 'new') as status"] = true
			case "total_dues", "payment_balance", "current_dues":
				selected["a.cache_total_dues as cache_total_dues"] = true
				selected["a.cache_payment_balance as cache_payment_balance"] = true
				selected["IFNULL(ad.flags, '') as admin_flags"] = true // needed for dues calc (guest!)
				selected["IFNULL(st.status, 'new') as status"] = true  // needed for dues calc
			case "due_date":
				selected["a.cache_due_date as cache_due_date"] = true
			case "registered":
				selected["a.created_at as created_at"] = true
			case "admin_comments":
				selected["IFNULL(ad.admin_comments, '') as admin_comments"] = true
			// custom field names
			case "name":
				selected["a.first_name as first_name"] = true
				selected["a.last_name as last_name"] = true
			case "address":
				selected["a.street as street"] = true
				selected["a.zip as zip"] = true
				selected["a.city as city"] = true
				selected["a.state as state"] = true
				selected["a.country as country"] = true
			case "contact":
				selected["a.email as email"] = true
				selected["a.phone as phone"] = true
				selected["a.telegram as telegram"] = true
				selected["a.spoken_languages as spoken_languages"] = true
			case "configuration":
				selected["a.registration_language as registration_language"] = true
				selected["a.flags as flags"] = true
				selected["a.options as options"] = true
				selected["a.packages as packages"] = true
			case "balances":
				selected["a.cache_total_dues as cache_total_dues"] = true
				selected["a.cache_payment_balance as cache_payment_balance"] = true
				selected["a.cache_open_balance as cache_open_balance"] = true
				selected["a.cache_due_date as cache_due_date"] = true
				selected["IFNULL(ad.flags, '') as admin_flags"] = true // needed for dues calc (guest!)
				selected["IFNULL(st.status, 'new') as status"] = true  // needed for dues calc
			case "all":
				selected["a.nickname as nickname"] = true
				selected["a.first_name as first_name"] = true
				selected["a.last_name as last_name"] = true
				selected["a.street as street"] = true
				selected["a.zip as zip"] = true
				selected["a.city as city"] = true
				selected["a.country as country"] = true
				selected["a.spoken_languages as spoken_languages"] = true
				selected["a.registration_language as registration_language"] = true
				selected["a.state as state"] = true
				selected["a.email as email"] = true
				selected["a.phone as phone"] = true
				selected["a.telegram as telegram"] = true
				selected["a.partner as partner"] = true
				selected["a.birthday as birthday"] = true
				selected["a.gender as gender"] = true
				selected["a.pronouns as pronouns"] = true
				selected["a.tshirt_size as tshirt_size"] = true
				selected["a.flags as flags"] = true
				selected["IFNULL(ad.flags, '') as admin_flags"] = true
				selected["a.options as options"] = true
				selected["a.packages as packages"] = true
				selected["a.user_comments as user_comments"] = true
				selected["IFNULL(st.status, 'new') as status"] = true
				selected["a.cache_total_dues as cache_total_dues"] = true
				selected["a.cache_payment_balance as cache_payment_balance"] = true
				selected["a.cache_open_balance as cache_open_balance"] = true
				selected["a.cache_due_date as cache_due_date"] = true
				selected["a.created_at as created_at"] = true
				selected["IFNULL(ad.admin_comments, '') as admin_comments"] = true
			default:
				// ignore
			}
		}
	}

	result := make([]string, len(selected))
	i := 0
	for k := range selected {
		result[i] = k
		i++
	}

	sort.Strings(result)
	return result
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

func (r *MysqlRepository) addSingleCondition(cond *attendee.AttendeeSearchSingleCriterion, params map[string]interface{}, idx int) string {
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
	if cond.Email != "" {
		query.WriteString(substringMatch("a.email", cond.Email, params, paramBaseName, &paramNo))
	}
	if cond.Telegram != "" {
		query.WriteString(substringMatch("a.telegram", cond.Telegram, params, paramBaseName, &paramNo))
	}
	query.WriteString(choiceMatch("a.spoken_languages", cond.SpokenLanguages, params, paramBaseName, &paramNo))
	query.WriteString(choiceMatch("a.registration_language", cond.RegistrationLanguage, params, paramBaseName, &paramNo))
	query.WriteString(choiceMatch("CONCAT(a.flags,IFNULL(ad.flags, ''))", cond.Flags, params, paramBaseName, &paramNo))
	query.WriteString(choiceMatch("a.options", cond.Options, params, paramBaseName, &paramNo))
	query.WriteString(choiceMatch("a.packages", cond.Packages, params, paramBaseName, &paramNo))
	if cond.UserComments != "" {
		query.WriteString(substringMatch("a.user_comments", cond.UserComments, params, paramBaseName, &paramNo))
	}
	if len(cond.Status) == 0 {
		query.WriteString("    AND ( IFNULL(st.status, 'new') <> 'deleted' )\n")
	} else {
		query.WriteString(safeStatusSliceMatch("IFNULL(st.status, 'new')", cond.Status))
	}
	query.WriteString(choiceMatch("IFNULL(ad.permissions, '')", cond.Permissions, params, paramBaseName, &paramNo))
	if cond.AdminComments != "" {
		query.WriteString(substringMatch("IFNULL(ad.admin_comments, '')", cond.AdminComments, params, paramBaseName, &paramNo))
	}
	if len(cond.AddInfo) > 0 {
		query.WriteString(r.addInfoConditions(cond.AddInfo, params, paramBaseName, &paramNo))
	}

	return query.String()
}

func (r *MysqlRepository) addInfoConditions(cond map[string]int8, params map[string]interface{}, paramBaseName string, idx *int) string {
	pName := fmt.Sprintf("%s_%d", paramBaseName, *idx)
	*idx++

	queryPart := strings.Builder{}

	for i, addInfoKey := range sortedKeySet(cond) {
		paramName := fmt.Sprintf("%s_%d", pName, i+1)
		wanted := cond[addInfoKey]
		if addInfoKey == "overdue" {
			todayIso := r.Now().Format(config.IsoDateFormat)
			if wanted == 0 {
				queryPart.WriteString(notOverdueCondition(todayIso, params, paramName))
			} else if wanted == 1 {
				queryPart.WriteString(overdueCondition(todayIso, params, paramName))
			}
		} else {
			if wanted == 0 {
				queryPart.WriteString(addInfoNotPresentCondition(addInfoKey, params, paramName))
			} else if wanted == 1 {
				queryPart.WriteString(addInfoPresentCondition(addInfoKey, params, paramName))
			}
		}
	}

	return queryPart.String()
}

func uintSliceMatch(field string, values []uint) string {
	mappedValues := make([]string, len(values))
	for i, v := range values {
		mappedValues[i] = fmt.Sprintf("%d", v)
	}
	return fmt.Sprintf("    AND ( %s IN (%s))\n", field, strings.Join(mappedValues, ","))
}

func safeStatusSliceMatch(field string, values []status.Status) string {
	allowedValues := config.AllowedStatusValues()
	mappedValues := make([]string, 0)
	for _, v := range values {
		for _, a := range allowedValues {
			// ensure no special characters
			if a == v {
				mappedValues = append(mappedValues, fmt.Sprintf("'%s'", v))
			}
		}
	}
	return fmt.Sprintf("    AND ( %s IN (%s) )\n", field, strings.Join(mappedValues, ","))
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
	keys := sortedKeySet(condition)

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

func sortedKeySet(condition map[string]int8) []string {
	keys := make([]string, len(condition))
	i := 0
	for k := range condition {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	return keys
}

func notOverdueCondition(todayIso string, params map[string]interface{}, paramName string) string {
	// a.cache_due_date = NULL => not considered overdue
	params[paramName] = todayIso
	return fmt.Sprintf("    AND ( STRCMP( IFNULL(a.cache_due_date,'0000-00-00'), @%s ) <= 0 )\n", paramName) +
		safeStatusSliceMatch("IFNULL(st.status, 'new')", []status.Status{status.Approved, status.PartiallyPaid})
}

func overdueCondition(todayIso string, params map[string]interface{}, paramName string) string {
	// a.cache_due_date = NULL => not considered overdue
	params[paramName] = todayIso
	return fmt.Sprintf("    AND ( STRCMP( IFNULL(a.cache_due_date,'0000-00-00'), @%s ) > 0 )\n", paramName) +
		safeStatusSliceMatch("IFNULL(st.status, 'new')", []status.Status{status.Approved, status.PartiallyPaid})
}

func addInfoNotPresentCondition(unsafeKey string, params map[string]interface{}, paramName string) string {
	params[paramName] = unsafeKey
	subquery := addInfoAreaCountSubquery(paramName)
	return fmt.Sprintf("    AND ( ( %s ) = 0 )\n", subquery)
}

func addInfoPresentCondition(unsafeKey string, params map[string]interface{}, paramName string) string {
	params[paramName] = unsafeKey
	subquery := addInfoAreaCountSubquery(paramName)
	return fmt.Sprintf("    AND ( ( %s ) > 0 )\n", subquery)
}

func addInfoAreaCountSubquery(paramName string) string {
	return fmt.Sprintf("SELECT COUNT(*) FROM att_additional_infos WHERE attendee_id = a.id AND area = @%s", paramName)
}
