package main

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"log"
	"os"
)

var AutoApply = false

// this program
// - reads config.yaml in this directory (not checked in)
// - queries IDP for identities of staff and directors group members
// - maps them to badge numbers via regsys API (TODO expose as admin endpoint)
// - queries regsys find endpoint for flags, status, nicks of all attendees
// - sets admin flags for staff / directors if missing
// - prints a message for non-attending flag holders and for flags that need to be cleared
func main() {
	config, err := loadValidatedConfig()
	if err != nil {
		log.Println("fatal: " + err.Error())
		os.Exit(1)
	}

	idpLookupResult, err := lookupUserIDs(config)
	if err != nil {
		log.Println("fatal: " + err.Error())
		os.Exit(2)
	}

	badgeLookupResult, err := lookupBadgeNumbers(idpLookupResult, config)
	if err != nil {
		log.Println("fatal: " + err.Error())
		os.Exit(3)
	}

	findResult, err := listAttendees(config)
	if err != nil {
		log.Println("fatal: " + err.Error())
		os.Exit(4)
	}

	log.Printf("we now have %d directors and %d staff and %d attendees", len(badgeLookupResult.DirectorBadges), len(badgeLookupResult.StaffBadges), len(findResult))

	for badgeNo, infos := range findResult {
		regStatus, shouldBeStaff := badgeLookupResult.StaffBadges[badgeNo]
		if regStatus == status.Cancelled {
			shouldBeStaff = false
		}
		if shouldBeStaff && !infos.Staff {
			log.Printf("id %d nick %s status %s should be staff", badgeNo, infos.Nickname, regStatus)
			if AutoApply {
				err := addAdminFlag(badgeNo, "staff", config)
				if err != nil {
					log.Printf("failed to add staff flag: %s", err.Error())
					os.Exit(5)
				}
			}
		} else if !shouldBeStaff && infos.Staff {
			log.Printf("id %d nick %s status %s should NOT be staff", badgeNo, infos.Nickname, regStatus)
		}

		regStatus, shouldBeDirector := badgeLookupResult.DirectorBadges[badgeNo]
		if regStatus == status.Cancelled {
			shouldBeDirector = false
		}
		if shouldBeDirector && !infos.Director {
			log.Printf("id %d nick %s status %s should be director", badgeNo, infos.Nickname, regStatus)
			if AutoApply {
				err := addAdminFlag(badgeNo, "director", config)
				if err != nil {
					log.Printf("failed to add director flag: %s", err.Error())
					os.Exit(5)
				}
			}
		} else if !shouldBeDirector && infos.Director {
			log.Printf("id %d nick %s status %s should NOT be director", badgeNo, infos.Nickname, regStatus)
		}
	}
}
