package attendeesrv

import (
	"context"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"math/rand"
	"strings"
	"time"
)

func (s *AttendeeServiceImplData) GenerateFakeRegistrations(ctx context.Context, count uint) error {
	for regNo := uint(0); regNo < count; regNo++ {
		attendee := fakeRegistration()
		id, err := database.GetRepository().AddAttendee(ctx, attendee)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Warn().Printf("failed to save attendee #%d with nickname %s - BAILING OUT", regNo+1, attendee.Nickname)
			return err
		}
		aulogging.Logger.Ctx(ctx).Info().Printf("successfully saved attendee #%d with badgeNo %d nickname %s", regNo+1, id, attendee.Nickname)
	}

	return nil
}

func fakeRegistration() *entity.Attendee {
	return &entity.Attendee{
		Nickname:             randomString(3, 40, 2),
		FirstName:            randomString(3, 30, 0),
		LastName:             randomString(3, 30, 0),
		Street:               randomString(3, 120, 0),
		Zip:                  randomString(3, 15, 0),
		City:                 randomString(3, 80, 0),
		Country:              "",
		Email:                "jsquirrel_github_9a6d@packetloss.de",
		Phone:                "12345",
		Birthday:             randomValidBirthday(),
		Pronouns:             "he/him",
		TshirtSize:           oneOf(config.AllowedTshirtSizes()),
		SpokenLanguages:      "," + randomSelection(config.AllowedSpokenLanguages(), 1, 5) + ",",
		RegistrationLanguage: oneOf(config.AllowedRegistrationLanguages()),
		Flags:                ",terms-accepted," + randomSelection([]string{"hc", "anon", "digi-book"}, 1, 3) + ",",
		Packages:             ",room-none,attendance," + oneOf([]string{"sponsor", "sponsor2", "tshirt"}) + ",",
		Options:              "," + randomSelection(config.AllowedOptions(), 1, 4) + ",",
		UserComments:         "",
		Identity:             randomString(10, 12, 0) + "_gen",
	}
}

func randomString(minlen int, maxlen int, maxNumSpecials int) string {
	text := ""
	normal := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	special := []rune("äöüÄÖÜß 松鼠วั")

	length := minlen + safeRandIntn(maxlen-minlen)
	specialcount := 0

	// always start and end in normal
	idx := rand.Intn(len(normal))
	text += fmt.Sprintf("%c", normal[idx])

	for i := 1; i < length-1; i++ {
		idx := rand.Intn(len(normal) + len(special))
		if idx >= len(normal) {
			specialcount++
			if specialcount > maxNumSpecials {
				idx := rand.Intn(len(normal))
				text += fmt.Sprintf("%c", normal[idx])
			} else {
				text += fmt.Sprintf("%c", special[idx-len(normal)])
			}
		} else {
			text += fmt.Sprintf("%c", normal[idx])
		}
	}

	idx = rand.Intn(len(normal))
	text += fmt.Sprintf("%c", normal[idx])

	return text
}

const isoDateFormat = "2006-01-02"

func randomValidBirthday() string {
	earliest, _ := time.Parse(isoDateFormat, config.EarliestBirthday())
	latest, _ := time.Parse(isoDateFormat, config.LatestBirthday())
	randUnix := earliest.Unix() + rand.Int63n(latest.Unix()-earliest.Unix())
	return time.Unix(randUnix, 0).Format(isoDateFormat)
}

func randomSelection(allowed []string, min int, max int) string {
	length := min + safeRandIntn(max-min)
	if length >= len(allowed) {
		return strings.Join(allowed, ",")
	}

	picked := make(map[int]bool)
	for i := range allowed {
		picked[i] = false
	}

	for i := 0; i < length; i++ {
		choice := safeRandIntn(len(allowed))
		for picked[choice] {
			choice = safeRandIntn(len(allowed))
		}
		picked[choice] = true
	}

	taken := make([]string, 0)
	for k, v := range picked {
		if v {
			taken = append(taken, allowed[k])
		}
	}
	return strings.Join(taken, ",")
}

func oneOf(allowed []string) string {
	picked := safeRandIntn(len(allowed))
	return allowed[picked]
}

func safeRandIntn(len int) int {
	if len > 0 {
		return rand.Intn(len)
	} else {
		return 0
	}
}
