package attendeesrv

import (
	"context"
	"fmt"

	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
)

func (s *AttendeeServiceImplData) RecordLimitChanges(ctx context.Context, deltas []*entity.Count) error {
	db := database.GetRepository()
	for _, delta := range deltas {
		if _, err := db.AddCount(ctx, delta); err != nil {
			return err
		}
	}
	return nil
}

func (s *AttendeeServiceImplData) IntroducesLimitOverrun(ctx context.Context, oldState *entity.Attendee, currentState *entity.Attendee, oldStatus status.Status, newStatus status.Status) ([]*entity.Count, error) {
	result := make([]*entity.Count, 0)

	packagesConfig := config.PackagesConfig()
	oldPackagesSelectedCountMap := choiceStrToMap(oldState.Packages, packagesConfig)
	currentPackagesSelectedCountMap := choiceStrToMap(currentState.Packages, packagesConfig)
	for key, conf := range packagesConfig {
		if conf.Limit > 0 {
			if currentPackagesSelectedCountMap[key] > 0 {
				// only adding / keeping a package can introduce overruns, so limit processing to this case
				oldPendingCount := oldPackagesSelectedCountMap[key] * pendingMultiplier(oldStatus)
				newPendingCount := currentPackagesSelectedCountMap[key] * pendingMultiplier(newStatus)

				oldAttendingCount := oldPackagesSelectedCountMap[key] * attendingMultiplier(oldStatus)
				newAttendingCount := currentPackagesSelectedCountMap[key] * attendingMultiplier(newStatus)

				currentAllocation, err := database.GetRepository().GetCount(ctx, entity.CountAreaPackage, key)
				if err != nil {
					return result, err
				}
				delta := entity.Count{
					Area:      entity.CountAreaPackage,
					Name:      key,
					Pending:   newPendingCount - oldPendingCount,
					Attending: newAttendingCount - oldAttendingCount,
				}
				if delta.Pending != 0 || delta.Attending != 0 {
					newPendingAllocation := currentAllocation.Pending + delta.Pending
					newAttendingAllocation := currentAllocation.Attending + delta.Attending

					if newPendingAllocation+newAttendingAllocation > conf.Limit {
						return result, fmt.Errorf("cannot allocate package '%s', allocation limit reached - please remove this package to continue: %w", key, IntroducesOverrun)
					}

					result = append(result, &delta)
				}
			}
		}
	}

	return result, nil
}

func pendingMultiplier(value status.Status) int {
	if value == status.New || value == status.Waiting {
		return 1
	} else {
		return 0
	}
}

func attendingMultiplier(value status.Status) int {
	if value == status.Approved || value == status.PartiallyPaid || value == status.Paid || value == status.CheckedIn {
		return 1
	} else {
		return 0
	}
}
