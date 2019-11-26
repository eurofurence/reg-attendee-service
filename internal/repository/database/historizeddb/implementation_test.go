package historizeddb

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/inmemorydb"
	"github.com/stretchr/testify/require"
	"testing"
)

func tstConstructCut() *HistorizingRepository {
	return Create(inmemorydb.Create()).(*HistorizingRepository)
}

func tstBuildValidAttendee() *entity.Attendee {
	return &entity.Attendee{
		Nickname:     "BlackCheetah",
		FirstName:    "Hans",
		LastName:     "Mustermann",
		Street:       "Teststraße 24",
		Zip:          "12345",
		City:         "Berlin",
		Country:      "DE",
		CountryBadge: "DE",
		State:        "Sachsen",
		Email:        "jsquirrel_github_9a6d@packetloss.de",
		Phone:        "+49-30-123",
		Telegram:     "@ihopethisuserdoesnotexist",
		Birthday:     "1998-11-23",
		Gender:       "other",
		Flags:        "anon,ev",
		Packages:     "room-none,attendance,stage,sponsor2",
		Options:      "music,suit",
		TshirtSize:   "XXL",
		UserComments: "this is a comment",
	}
}

func TestHistorizesUpdateCorrectly(t *testing.T) {
	docs.Description("check that historizing changes works as expected")
	cut := tstConstructCut()
	cut.Open()
	cut.Migrate()

	a := tstBuildValidAttendee()
	id, err := cut.AddAttendee(context.TODO(), a)
	require.Nil(t, err, "unexpected error during add")

	a.Nickname = "WhiteCheetah"
	a.State = ""
	a.UserComments = "this change should be invisible"
	err = cut.UpdateAttendee(context.TODO(), a)
	require.Nil(t, err, "unexpected error during update")

	expectedDiff := "modified: .Nickname = \"BlackCheetah\"\nmodified: .State = \"Sachsen\"\n"
	actualDiff, err := cut.wrappedRepository.(*inmemorydb.InMemoryRepository).GetHistoryById(context.TODO(), id + 1)
	require.Nil(t, err, "unexpected error during history access")
	require.Equal(t, expectedDiff, actualDiff.Diff)

	cut.Close()
}

func TestDirectRecordHistoryFails(t *testing.T) {
	docs.Description("check that trying to directly write to the history fails")
	cut := tstConstructCut()
	cut.Open()

	h := &entity.History{}
	err := cut.RecordHistory(context.TODO(), h)
	require.NotNil(t, err)

	cut.Close()
}
