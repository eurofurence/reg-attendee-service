package historizeddb

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/inmemorydb"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

func tstConstructCut() *HistorizingRepository {
	return Create(inmemorydb.Create()).(*HistorizingRepository)
}

func tstBuildValidAttendee() *entity.Attendee {
	return &entity.Attendee{
		Nickname:             "BlackCheetah",
		FirstName:            "Hans",
		LastName:             "Mustermann",
		Street:               "Teststraße 24",
		Zip:                  "12345",
		City:                 "Berlin",
		Country:              "DE",
		State:                "Sachsen",
		Email:                "jsquirrel_github_9a6d@packetloss.de",
		Phone:                "+49-30-123",
		Telegram:             "@ihopethisuserdoesnotexist",
		Birthday:             "1998-11-23",
		Gender:               "other",
		SpokenLanguages:      ",de_DE,en_US,",
		RegistrationLanguage: ",en_US,",
		Flags:                ",anon,ev,",
		Packages:             ",room-none,attendance,stage,sponsor2,",
		Options:              ",music,suit,",
		TshirtSize:           "XXL",
		UserComments:         "this is a comment",
	}
}

func tstBuildValidAdminInfo1() *entity.AdminInfo {
	return &entity.AdminInfo{
		Model:         gorm.Model{ID: 1},
		Flags:         "",
		Permissions:   "admin",
		AdminComments: "first",
	}
}

func tstBuildValidAdminInfo2() *entity.AdminInfo {
	return &entity.AdminInfo{
		Model:         gorm.Model{ID: 1},
		Flags:         "something",
		Permissions:   "",
		AdminComments: "second",
	}
}

func TestHistorizesUpdateCorrectly(t *testing.T) {
	docs.Description("check that historizing attendee changes works as expected")
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
	actualDiff, err := cut.wrappedRepository.(*inmemorydb.InMemoryRepository).GetHistoryById(context.TODO(), id+1)
	require.Nil(t, err, "unexpected error during history access")
	require.Equal(t, expectedDiff, actualDiff.Diff)

	cut.Close()
}

func TestHistorizesAdminInfoChangesCorrectly(t *testing.T) {
	docs.Description("check that historizing admin info changes works as expected")
	cut := tstConstructCut()
	cut.Open()
	cut.Migrate()

	orig := tstBuildValidAdminInfo1()
	err := cut.WriteAdminInfo(context.TODO(), orig)
	require.Nil(t, err, "unexpected error during initial add")

	change := tstBuildValidAdminInfo2()
	err = cut.WriteAdminInfo(context.TODO(), change)
	require.Nil(t, err, "unexpected error during update")

	expectedDiff1 := "modified: .Permissions = \"\"\n"
	actualDiff1, err := cut.wrappedRepository.(*inmemorydb.InMemoryRepository).GetHistoryById(context.TODO(), 1)
	require.Nil(t, err, "unexpected error during history access 1")
	require.Equal(t, expectedDiff1, actualDiff1.Diff)

	expectedDiff2 := "modified: .Flags = \"\"\nmodified: .Permissions = \"admin\"\n"
	actualDiff2, err := cut.wrappedRepository.(*inmemorydb.InMemoryRepository).GetHistoryById(context.TODO(), 2)
	require.Nil(t, err, "unexpected error during history access 2")
	require.Equal(t, expectedDiff2, actualDiff2.Diff)

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
