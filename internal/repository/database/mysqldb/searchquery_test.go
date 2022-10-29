package mysqldb

import (
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEmptySearchQuery(t *testing.T) {
	spec := &attendee.AttendeeSearchCriteria{}

	actualParams := make(map[string]interface{})
	actualQuery := constructAttendeeSearchQuery(spec, actualParams)

	expectedParams := map[string]interface{}{}
	expectedQuery := `SELECT * FROM attendees a WHERE (
  (0 = 1)
) ORDER BY a.id `

	require.Equal(t, expectedQuery, actualQuery)
	require.EqualValues(t, expectedParams, actualParams)
}

func TestTwoFullSearchQueries(t *testing.T) {
	spec := &attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{
				Ids:          []uint{22, 33, 44},
				Nickname:     "*chee*",
				Name:         "John D*e",
				Address:      "Berlin",
				Country:      "DE",
				CountryBadge: "UK",
				Email:        "ee*ee@ff*ff",
				Telegram:     "@abc",
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
			},
			{
				Ids:          []uint{23, 34, 45},
				Nickname:     "small*bird",
				Name:         "Johnny",
				Address:      "Berlin",
				Country:      "CH",
				CountryBadge: "IT",
				Email:        "gg@hh",
				Telegram:     "@def",
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
			},
		},
		MinId:     1,
		MaxId:     400,
		SortBy:    "name",
		SortOrder: "descending",
	}

	actualParams := make(map[string]interface{})
	actualQuery := constructAttendeeSearchQuery(spec, actualParams)

	str := ""
	for k, v := range actualParams {
		str = str + `"` + k + `": "` + fmt.Sprintf("%v", v) + `",
`
	}

	expectedParams := map[string]interface{}{
		"param_0_1":  uint(1),
		"param_0_2":  uint(400),
		"param_1_1":  "%chee%",
		"param_1_2":  "John D%e",
		"param_1_3":  "%Berlin%",
		"param_1_4":  "DE",
		"param_1_5":  "UK",
		"param_1_6":  "%ee%ee@ff%ff%",
		"param_1_7":  "%@abc%",
		"param_1_8":  "%,flagone,%",
		"param_1_9":  "%,flagzero,%",
		"param_1_10": "%,optone,%",
		"param_1_11": "%,optzero,%",
		"param_1_12": "%,pkgone,%",
		"param_1_13": "%,pkgzero,%",
		"param_1_14": "%user%comments%",
		"param_2_1":  "small%bird",
		"param_2_2":  "Johnny",
		"param_2_3":  "%Berlin%",
		"param_2_4":  "CH",
		"param_2_5":  "IT",
		"param_2_6":  "%gg@hh%",
		"param_2_7":  "%@def%",
		"param_2_8":  "%,fone,%",
		"param_2_9":  "%,fzero,%",
		"param_2_10": "%,oone,%",
		"param_2_11": "%,ozero,%",
		"param_2_12": "%,pone,%",
		"param_2_13": "%,pzero,%",
		"param_2_14": "%more user comments%",
	}
	expectedQuery := `SELECT * FROM attendees a WHERE (
  (0 = 1)
  OR
  (
    (1 = 1)
    AND ( a.id IN (22,33,44))
    AND ( LOWER(a.nickname) LIKE LOWER( @param_1_1 ) )
    AND ( LOWER(CONCAT(a.first_name, ' ', a.last_name)) LIKE LOWER( @param_1_2 ) )
    AND ( LOWER(CONCAT(a.street, ' ', a.zip, ' ', a.city, ' ', a.state)) LIKE LOWER( @param_1_3 ) )
    AND ( LOWER(a.country) = LOWER( @param_1_4 ) )
    AND ( LOWER(a.country_badge) = LOWER( @param_1_5 ) )
    AND ( LOWER(a.email) LIKE LOWER( @param_1_6 ) )
    AND ( LOWER(a.telegram) LIKE LOWER( @param_1_7 ) )
    AND ( a.flags LIKE @param_1_8 )
    AND ( a.flags NOT LIKE @param_1_9 )
    AND ( a.options LIKE @param_1_10 )
    AND ( a.options NOT LIKE @param_1_11 )
    AND ( a.packages LIKE @param_1_12 )
    AND ( a.packages NOT LIKE @param_1_13 )
    AND ( LOWER(a.user_comments) LIKE LOWER( @param_1_14 ) )
  )
  OR
  (
    (1 = 1)
    AND ( a.id IN (23,34,45))
    AND ( LOWER(a.nickname) LIKE LOWER( @param_2_1 ) )
    AND ( LOWER(CONCAT(a.first_name, ' ', a.last_name)) LIKE LOWER( @param_2_2 ) )
    AND ( LOWER(CONCAT(a.street, ' ', a.zip, ' ', a.city, ' ', a.state)) LIKE LOWER( @param_2_3 ) )
    AND ( LOWER(a.country) = LOWER( @param_2_4 ) )
    AND ( LOWER(a.country_badge) = LOWER( @param_2_5 ) )
    AND ( LOWER(a.email) LIKE LOWER( @param_2_6 ) )
    AND ( LOWER(a.telegram) LIKE LOWER( @param_2_7 ) )
    AND ( a.flags LIKE @param_2_8 )
    AND ( a.flags NOT LIKE @param_2_9 )
    AND ( a.options LIKE @param_2_10 )
    AND ( a.options NOT LIKE @param_2_11 )
    AND ( a.packages LIKE @param_2_12 )
    AND ( a.packages NOT LIKE @param_2_13 )
    AND ( LOWER(a.user_comments) LIKE LOWER( @param_2_14 ) )
  )
) AND a.id >= @param_0_1 AND a.id <= @param_0_2 ORDER BY CONCAT(a.first_name, ' ', a.last_name) DESC`

	require.Equal(t, expectedQuery, actualQuery)
	require.EqualValues(t, expectedParams, actualParams)
}
