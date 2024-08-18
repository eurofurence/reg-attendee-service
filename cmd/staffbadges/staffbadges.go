package main

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"log"
	"os"
	"sort"
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

	log.Printf("we now have %d directors and %d staff (still includes cancelled) and %d attendees", len(badgeLookupResult.DirectorBadges), len(badgeLookupResult.StaffBadges), len(findResult))

	// sort by badgeNo
	badgeNumbers := make([]uint, 0, len(findResult))
	for badgeNo, _ := range findResult {
		badgeNumbers = append(badgeNumbers, badgeNo)
	}
	sort.Slice(badgeNumbers, func(i, j int) bool { return badgeNumbers[i] < badgeNumbers[j] })

	// remove cancelled in separate pre-processing step
	for _, badgeNo := range badgeNumbers {
		infos := findResult[badgeNo]

		regStatus, present := badgeLookupResult.StaffBadges[badgeNo]
		if present && regStatus == status.Cancelled {
			log.Printf("cancelled staff id %d nick %s", badgeNo, infos.Nickname)
			delete(badgeLookupResult.StaffBadges, badgeNo)
		}

		regStatus, present = badgeLookupResult.DirectorBadges[badgeNo]
		if present && regStatus == status.Cancelled {
			log.Printf("cancelled director id %d nick %s", badgeNo, infos.Nickname)
			delete(badgeLookupResult.DirectorBadges, badgeNo)
		}
	}

	log.Printf("we now have %d directors and %d staff and %d attendees", len(badgeLookupResult.DirectorBadges), len(badgeLookupResult.StaffBadges), len(findResult))

	for _, badgeNo := range badgeNumbers {
		infos := findResult[badgeNo]

		regStatus, shouldBeStaff := badgeLookupResult.StaffBadges[badgeNo]
		if shouldBeStaff && !infos.Staff {
			log.Printf("id %d nick %s status %s should be staff", badgeNo, infos.Nickname, regStatus)
			if AutoApply {
				err := addAdminFlag(badgeNo, "staff", config)
				if err != nil {
					log.Printf("failed to add staff flag: %s", err.Error())
					os.Exit(5)
				}
				log.Printf("added staff flag for id %d nick %s", badgeNo, infos.Nickname)
			}
		} else if !shouldBeStaff && infos.Staff {
			log.Printf("id %d nick %s status %s should NOT be staff - removal is manual", badgeNo, infos.Nickname, regStatus)
		}

		regStatus, shouldBeDirector := badgeLookupResult.DirectorBadges[badgeNo]
		if shouldBeDirector && !infos.Director {
			log.Printf("id %d nick %s status %s should be director", badgeNo, infos.Nickname, regStatus)
			if AutoApply {
				err := addAdminFlag(badgeNo, "director", config)
				if err != nil {
					log.Printf("failed to add director flag: %s", err.Error())
					os.Exit(5)
				}
				log.Printf("added director flag for id %d nick %s", badgeNo, infos.Nickname)
			}
		} else if !shouldBeDirector && infos.Director {
			log.Printf("id %d nick %s status %s should NOT be director - removal is manual", badgeNo, infos.Nickname, regStatus)
		}
	}
}
