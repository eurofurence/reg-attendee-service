package mysqldb

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func tstConstructClassUnderTest() *MysqlRepository {
	mockNow, _ := time.Parse(config.IsoDateFormat, "2020-12-23")
	return &MysqlRepository{
		Now: func() time.Time { return mockNow },
	}
}

func TestEmptySearchQuery(t *testing.T) {
	cut := tstConstructClassUnderTest()
	spec := &attendee.AttendeeSearchCriteria{}

	actualParams := make(map[string]interface{})
	actualQuery := cut.constructAttendeeSearchQuery(context.Background(), spec, actualParams)

	expectedParams := map[string]interface{}{
		"param_force_named_query_detection": 1,
	}
	expectedQuery := `SELECT IFNULL(ad.admin_comments, '') as admin_comments, IFNULL(ad.flags, '') as admin_flags, IFNULL(st.status, 'new') as status, a.birthday as birthday, a.cache_due_date as cache_due_date, a.cache_open_balance as cache_open_balance, a.cache_payment_balance as cache_payment_balance, a.cache_total_dues as cache_total_dues, a.country as country, a.created_at as created_at, a.email as email, a.first_name as first_name, a.flags as flags, a.id as id, a.last_name as last_name, a.nickname as nickname, a.options as options, a.packages as packages, a.pronouns as pronouns, a.spoken_languages as spoken_languages, a.telegram as telegram, a.tshirt_size as tshirt_size, a.user_comments as user_comments 
FROM att_attendees AS a 
  LEFT JOIN att_admin_infos AS ad ON ad.id = a.id 
  LEFT JOIN (  SELECT sc.attendee_id AS attendee_id,         ( SELECT sc2.status FROM att_status_changes AS sc2 WHERE sc2.id = max(sc.id) ) AS status  FROM att_status_changes AS sc  GROUP BY sc.attendee_id  ) AS st ON st.attendee_id = a.id 
WHERE (
  (0 = @param_force_named_query_detection)
) ORDER BY a.id `

	require.Equal(t, expectedQuery, actualQuery)
	require.EqualValues(t, expectedParams, actualParams)
}

func TestTwoFullSearchQueries(t *testing.T) {
	cut := tstConstructClassUnderTest()
	spec := &attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{
				Ids:      []uint{22, 33, 44},
				Nickname: "*chee*",
				Name:     "John D*e",
				Address:  "Berlin",
				Country:  "DE",
				Email:    "ee*ee@ff*ff",
				Telegram: "@abc",
				SpokenLanguages: map[string]int8{
					"en-US": 0,
					"de-DE": 1,
				},
				RegistrationLanguage: map[string]int8{
					"en-US": 0,
					"de-DE": 1,
				},
				Flags: map[string]int8{
					"flagone":  1,
					"flagzero": 0,
				},
				Options: map[string]int8{
					"optone":  1,
					"optzero": 0,
				},
				Packages: map[string]int8{
					"pkgone":  1,
					"pkgzero": 0,
				},
				UserComments: "user*comments",
				AddInfo: map[string]int8{
					"overdue":       0,
					"sponsor-items": 0,
				},
			},
			{
				Ids:      []uint{23, 34, 45},
				Nickname: "small*bird",
				Name:     "Johnny",
				Address:  "Berlin",
				Country:  "CH",
				Email:    "gg@hh",
				Telegram: "@def",
				SpokenLanguages: map[string]int8{
					"en-GB": 0,
					"de-AT": 1,
				},
				RegistrationLanguage: map[string]int8{
					"en-GB": 0,
					"de-AT": 1,
				},
				Flags: map[string]int8{
					"fone":  1,
					"fzero": 0,
				},
				Options: map[string]int8{
					"oone":  1,
					"ozero": 0,
				},
				Packages: map[string]int8{
					"pone":  1,
					"pzero": 0,
				},
				UserComments: "more user comments",
				AddInfo: map[string]int8{
					"overdue":       1,
					"sponsor-items": 1,
				},
				BirthdayFrom: "1970-10-24",
				BirthdayTo:   "1980-12-24",
			},
		},
		MinId:      1,
		MaxId:      400,
		SortBy:     "name",
		SortOrder:  "descending",
		FillFields: []string{"configuration", "balances", "pronouns"},
	}

	actualParams := make(map[string]interface{})
	actualQuery := cut.constructAttendeeSearchQuery(context.Background(), spec, actualParams)

	str := ""
	for k, v := range actualParams {
		str = str + `"` + k + `": "` + fmt.Sprintf("%v", v) + `",
`
	}

	expectedParams := map[string]interface{}{
		"param_force_named_query_detection": 1,
		"param_0_1":                         uint(1),
		"param_0_2":                         uint(400),
		"param_1_1":                         "%chee%",
		"param_1_2":                         "John D%e",
		"param_1_3":                         "%Berlin%",
		"param_1_4":                         "DE",
		"param_1_5":                         "%ee%ee@ff%ff%",
		"param_1_6":                         "%@abc%",
		"param_1_7":                         "%,de-DE,%",
		"param_1_8":                         "%,en-US,%",
		"param_1_9":                         "%,de-DE,%",
		"param_1_10":                        "%,en-US,%",
		"param_1_11":                        "%,flagone,%",
		"param_1_12":                        "%,flagzero,%",
		"param_1_13":                        "%,optone,%",
		"param_1_14":                        "%,optzero,%",
		"param_1_15":                        "%,pkgone,%",
		"param_1_16":                        "%,pkgzero,%",
		"param_1_17":                        "%user%comments%",
		"param_1_18_1":                      "2020-12-23",
		"param_1_18_2":                      "sponsor-items",
		"param_2_1":                         "small%bird",
		"param_2_2":                         "Johnny",
		"param_2_20":                        "1980-12-24",
		"param_2_3":                         "%Berlin%",
		"param_2_4":                         "CH",
		"param_2_5":                         "%gg@hh%",
		"param_2_6":                         "%@def%",
		"param_2_7":                         "%,de-AT,%",
		"param_2_8":                         "%,en-GB,%",
		"param_2_9":                         "%,de-AT,%",
		"param_2_10":                        "%,en-GB,%",
		"param_2_11":                        "%,fone,%",
		"param_2_12":                        "%,fzero,%",
		"param_2_13":                        "%,oone,%",
		"param_2_14":                        "%,ozero,%",
		"param_2_15":                        "%,pone,%",
		"param_2_16":                        "%,pzero,%",
		"param_2_17":                        "%more user comments%",
		"param_2_18_1":                      "2020-12-23",
		"param_2_18_2":                      "sponsor-items",
		"param_2_19":                        "1970-10-24",
	}
	expectedQuery := `SELECT IFNULL(ad.flags, '') as admin_flags, IFNULL(st.status, 'new') as status, a.cache_due_date as cache_due_date, a.cache_open_balance as cache_open_balance, a.cache_payment_balance as cache_payment_balance, a.cache_total_dues as cache_total_dues, a.flags as flags, a.id as id, a.options as options, a.packages as packages, a.pronouns as pronouns, a.registration_language as registration_language 
FROM att_attendees AS a 
  LEFT JOIN att_admin_infos AS ad ON ad.id = a.id 
  LEFT JOIN (  SELECT sc.attendee_id AS attendee_id,         ( SELECT sc2.status FROM att_status_changes AS sc2 WHERE sc2.id = max(sc.id) ) AS status  FROM att_status_changes AS sc  GROUP BY sc.attendee_id  ) AS st ON st.attendee_id = a.id 
WHERE (
  (0 = @param_force_named_query_detection)
  OR
  (
    (1 = 1)
    AND ( a.id IN (22,33,44))
    AND ( LOWER(a.nickname) LIKE LOWER( @param_1_1 ) )
    AND ( LOWER(CONCAT(a.first_name, ' ', a.last_name)) LIKE LOWER( @param_1_2 ) )
    AND ( LOWER(CONCAT(a.street, ' ', a.zip, ' ', a.city, ' ', a.state)) LIKE LOWER( @param_1_3 ) )
    AND ( LOWER(a.country) = LOWER( @param_1_4 ) )
    AND ( LOWER(a.email) LIKE LOWER( @param_1_5 ) )
    AND ( LOWER(a.telegram) LIKE LOWER( @param_1_6 ) )
    AND ( a.spoken_languages LIKE @param_1_7 )
    AND ( a.spoken_languages NOT LIKE @param_1_8 )
    AND ( a.registration_language LIKE @param_1_9 )
    AND ( a.registration_language NOT LIKE @param_1_10 )
    AND ( CONCAT(a.flags,IFNULL(ad.flags, '')) LIKE @param_1_11 )
    AND ( CONCAT(a.flags,IFNULL(ad.flags, '')) NOT LIKE @param_1_12 )
    AND ( a.options LIKE @param_1_13 )
    AND ( a.options NOT LIKE @param_1_14 )
    AND ( a.packages LIKE @param_1_15 )
    AND ( a.packages NOT LIKE @param_1_16 )
    AND ( LOWER(a.user_comments) LIKE LOWER( @param_1_17 ) )
    AND ( IFNULL(st.status, 'new') <> 'deleted' )
    AND ( STRCMP( IFNULL(a.cache_due_date,'9999-99-99'), @param_1_18_1 ) >= 0 )
    AND ( IFNULL(st.status, 'new') IN ('approved','partially paid') )
    AND ( ( SELECT COUNT(*) FROM att_additional_infos WHERE attendee_id = a.id AND area = @param_1_18_2 ) = 0 )
  )
  OR
  (
    (1 = 1)
    AND ( a.id IN (23,34,45))
    AND ( LOWER(a.nickname) LIKE LOWER( @param_2_1 ) )
    AND ( LOWER(CONCAT(a.first_name, ' ', a.last_name)) LIKE LOWER( @param_2_2 ) )
    AND ( LOWER(CONCAT(a.street, ' ', a.zip, ' ', a.city, ' ', a.state)) LIKE LOWER( @param_2_3 ) )
    AND ( LOWER(a.country) = LOWER( @param_2_4 ) )
    AND ( LOWER(a.email) LIKE LOWER( @param_2_5 ) )
    AND ( LOWER(a.telegram) LIKE LOWER( @param_2_6 ) )
    AND ( a.spoken_languages LIKE @param_2_7 )
    AND ( a.spoken_languages NOT LIKE @param_2_8 )
    AND ( a.registration_language LIKE @param_2_9 )
    AND ( a.registration_language NOT LIKE @param_2_10 )
    AND ( CONCAT(a.flags,IFNULL(ad.flags, '')) LIKE @param_2_11 )
    AND ( CONCAT(a.flags,IFNULL(ad.flags, '')) NOT LIKE @param_2_12 )
    AND ( a.options LIKE @param_2_13 )
    AND ( a.options NOT LIKE @param_2_14 )
    AND ( a.packages LIKE @param_2_15 )
    AND ( a.packages NOT LIKE @param_2_16 )
    AND ( LOWER(a.user_comments) LIKE LOWER( @param_2_17 ) )
    AND ( IFNULL(st.status, 'new') <> 'deleted' )
    AND ( STRCMP( IFNULL(a.cache_due_date,'9999-99-99'), @param_2_18_1 ) < 0 )
    AND ( IFNULL(st.status, 'new') IN ('approved','partially paid') )
    AND ( ( SELECT COUNT(*) FROM att_additional_infos WHERE attendee_id = a.id AND area = @param_2_18_2 ) > 0 )
    AND ( STRCMP(a.birthday,@param_2_19) >= 0 )
    AND ( STRCMP(a.birthday,@param_2_20) <= 0 )
  )
) AND a.id >= @param_0_1 AND a.id <= @param_0_2 ORDER BY CONCAT(a.first_name, ' ', a.last_name) DESC`

	require.Equal(t, expectedQuery, actualQuery)
	require.EqualValues(t, expectedParams, actualParams)
}
