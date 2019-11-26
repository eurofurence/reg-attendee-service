package inmemorydb

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"testing"
)

var (
	cut *InMemoryRepository
)

func TestMain(m *testing.M) {
	cut = &InMemoryRepository{}
	cut.Open()
	code := m.Run()
	cut.Close()
	os.Exit(code)
}

func TestOpenClose(t *testing.T) {
	docs.Description("low level test for Open() and Close()")
	cut2 := &InMemoryRepository{}
	cut2.Open()
	require.NotNil(t, cut2.attendees)
	cut2.Close()
	require.Nil(t, cut2.attendees)
}

func TestAddAttendee(t *testing.T) {
	docs.Description("it should be possible to add an attendee and then retrieve it again")
	att := &entity.Attendee{}
	newId, err := cut.AddAttendee(context.TODO(), att)
	require.Nil(t, err, "unexpected error during add")

	att2, err := cut.GetAttendeeById(context.TODO(), newId)
	require.Nil(t, err, "unexpected error during get")
	require.EqualValues(t, *att, *att2, "comparison failure")
}

func TestGetAttendeeNotFound(t *testing.T) {
	docs.Description("retrieving a nonexistent attendee should fail")
	att, err := cut.GetAttendeeById(context.TODO(), 0)
	require.NotNil(t, err, "no error occurred, although it should have")
	require.Equal(t, "cannot get attendee 0 - not present", err.Error(), "unexpected error message")
	require.Equal(t, uint(0), att.ID, "ID should still be at its initial value")
}

func TestUpdateAttendee(t *testing.T) {
	docs.Description("it should be possible to update an attendee")
	att := &entity.Attendee{}
	att.Nickname = "something"
	newId, err := cut.AddAttendee(context.TODO(), att)
	require.Nil(t, err, "unexpected error during add")

	att2, err := cut.GetAttendeeById(context.TODO(), newId)
	require.Nil(t, err, "unexpected error during get")
	att2.Nickname = "somethingelse"
	require.Equal(t, newId, att2.ID, "unexpected difference in id after get")

	err = cut.UpdateAttendee(context.TODO(), att2)
	require.Nil(t, err, "unexpected error during update")
	require.Equal(t, "somethingelse", cut.attendees[newId].Nickname, "changed value not recorded in db")
}

func TestUpdateAttendeeNotFound(t *testing.T) {
	docs.Description("updating a nonexistent attendee should fail")
	att := &entity.Attendee{}
	err := cut.UpdateAttendee(context.TODO(), att)
	require.NotNil(t, err, "no error occurred, although it should have")
	require.Equal(t, "cannot update attendee 0 - not present", err.Error(), "unexpected error message")
	require.Equal(t, uint(0), att.ID, "ID should still be at its initial value")
}
