package inmemorydb

import (
	"context"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/dbrepo"
	"gorm.io/gorm"
	"sort"
	"sync/atomic"
	"time"
)

type InMemoryRepository struct {
	addInfo       map[uint]map[string]*entity.AdditionalInfo
	adminInfo     map[uint]*entity.AdminInfo
	attendees     map[uint]*entity.Attendee
	bans          map[uint]*entity.Ban
	statusChanges map[uint][]entity.StatusChange
	history       map[uint]*entity.History
	idSequence    uint32
	Now           func() time.Time
}

func Create() dbrepo.Repository {
	return &InMemoryRepository{
		Now: time.Now,
	}
}

func (r *InMemoryRepository) Open() error {
	r.addInfo = make(map[uint]map[string]*entity.AdditionalInfo)
	r.adminInfo = make(map[uint]*entity.AdminInfo)
	r.attendees = make(map[uint]*entity.Attendee)
	r.bans = make(map[uint]*entity.Ban)
	r.statusChanges = make(map[uint][]entity.StatusChange)
	r.history = make(map[uint]*entity.History)
	return nil
}

func (r *InMemoryRepository) Close() {
	r.addInfo = nil
	r.adminInfo = nil
	r.attendees = nil
	r.bans = nil
	r.statusChanges = nil
	r.history = nil
}

func (r *InMemoryRepository) Migrate() error {
	// nothing to do
	return nil
}

// --- attendee ---

func (r *InMemoryRepository) AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error) {
	newId := uint(atomic.AddUint32(&r.idSequence, 1))
	a.ID = newId

	// copy the attendee, so later modifications won't also modify it in the simulated db
	copiedAttendee := *a
	copiedAttendee.CreatedAt = time.Now()
	r.attendees[newId] = &copiedAttendee
	return newId, nil
}

func (r *InMemoryRepository) UpdateAttendee(ctx context.Context, a *entity.Attendee) error {
	if _, ok := r.attendees[a.ID]; ok {
		// allow deleted because the admin ui does

		// copy the attendee, so later modifications won't also modify it in the simulated db
		copiedAttendee := *a
		copiedAttendee.UpdatedAt = time.Now()
		r.attendees[a.ID] = &copiedAttendee
		return nil
	} else {
		return fmt.Errorf("cannot update attendee %d - not present", a.ID)
	}
}

func (r *InMemoryRepository) GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error) {
	if att, ok := r.attendees[id]; ok {
		// allow deleted so history and undelete work

		// copy the attendee, so later modifications won't also modify it in the simulated db
		copiedAttendee := *att
		return &copiedAttendee, nil
	} else {
		return &entity.Attendee{}, fmt.Errorf("cannot get attendee %d - not present", id)
	}
}

func (r *InMemoryRepository) SoftDeleteAttendeeById(ctx context.Context, id uint) error {
	if att, ok := r.attendees[id]; ok {
		att.DeletedAt = gorm.DeletedAt{
			Time:  r.Now(),
			Valid: true,
		}
		return nil
	} else {
		return fmt.Errorf("cannot delete attendee %d - not present", id)
	}
}

func (r *InMemoryRepository) UndeleteAttendeeById(ctx context.Context, id uint) error {
	if att, ok := r.attendees[id]; ok {
		att.DeletedAt = gorm.DeletedAt{
			Time:  r.Now(),
			Valid: false,
		}
		return nil
	} else {
		return fmt.Errorf("cannot delete attendee %d - not present", id)
	}
}

func (r *InMemoryRepository) CountAttendeesByNicknameZipEmail(ctx context.Context, nickname string, zip string, email string) (int64, error) {
	var count int64
	for _, v := range r.attendees {
		// count deleted because the unique index in the db will
		if nickname == v.Nickname && zip == v.Zip && email == v.Email {
			count++
		}
	}
	return count, nil
}

func (r *InMemoryRepository) CountAttendeesByIdentity(ctx context.Context, identity string) (int64, error) {
	var count int64
	for _, v := range r.attendees {
		// count deleted because the unique index in the db will
		if identity == v.Identity {
			count++
		}
	}
	return count, nil
}

func (r *InMemoryRepository) MaxAttendeeId(ctx context.Context) (uint, error) {
	var max uint
	for _, v := range r.attendees {
		if v.ID > max {
			max = v.ID
		}
	}
	return max, nil
}

// --- attendee search ---

func (r *InMemoryRepository) FindAttendees(ctx context.Context, criteria *attendee.AttendeeSearchCriteria) ([]*entity.AttendeeQueryResult, error) {
	resultIds := make([]uint, 0)
	for id, a := range r.attendees {
		adm, _ := r.GetAdminInfoByAttendeeId(ctx, a.ID)
		sc, _ := r.GetLatestStatusChangeByAttendeeId(ctx, a.ID)
		addInfs := r.GetAllAdditionalInfoOrEmptyMap(ctx, a.ID)
		if r.matchesCriteria(criteria, a, adm, sc, addInfs) {
			resultIds = append(resultIds, id)
		}
	}

	sort.Slice(resultIds, r.lessFunction(criteria.SortBy, criteria.SortOrder, resultIds))

	resultLen := len(resultIds)
	if criteria.NumResults > 0 && resultLen > int(criteria.NumResults) {
		resultLen = int(criteria.NumResults)
	}

	result := make([]*entity.AttendeeQueryResult, resultLen)
	for i, aid := range resultIds {
		if i < resultLen {
			copiedAttendee := *(r.attendees[aid])
			adminInfo, _ := r.GetAdminInfoByAttendeeId(ctx, aid)
			latestStatus, _ := r.GetLatestStatusChangeByAttendeeId(ctx, aid)
			copiedResult := entity.AttendeeQueryResult{
				Attendee:      copiedAttendee,
				Status:        latestStatus.Status,
				AdminComments: adminInfo.AdminComments,
				AdminFlags:    adminInfo.Flags,
			}
			result[i] = &copiedResult
		}
	}

	return result, nil
}

func (r *InMemoryRepository) lessFunction(sortBy string, sortOrder string, matchingIds []uint) func(i, j int) bool {
	return func(i, j int) bool {
		a1 := r.attendees[matchingIds[i]]
		a2 := r.attendees[matchingIds[j]]
		switch sortBy {
		case "status":
			// TODO status lookup and sort by it
			return lessFunctionId(a1, a2, sortOrder)
		case "nickname":
			return lessFunctionString(a1, a2, func(a *entity.Attendee) string { return a.Nickname }, sortOrder)
		case "birthday":
			return lessFunctionString(a1, a2, func(a *entity.Attendee) string { return a.Birthday }, sortOrder)
		case "email":
			return lessFunctionString(a1, a2, func(a *entity.Attendee) string { return a.Email }, sortOrder)
		case "name":
			return lessFunctionString(a1, a2, func(a *entity.Attendee) string { return a.FirstName + " " + a.LastName }, sortOrder)
		case "zip":
			return lessFunctionString(a1, a2, func(a *entity.Attendee) string { return a.Zip }, sortOrder)
		case "city":
			return lessFunctionString(a1, a2, func(a *entity.Attendee) string { return a.City }, sortOrder)
		case "country":
			return lessFunctionString(a1, a2, func(a *entity.Attendee) string { return a.Country }, sortOrder)
		default:
			return lessFunctionId(a1, a2, sortOrder)
		}
	}
}

func lessFunctionId(a1 *entity.Attendee, a2 *entity.Attendee, sortOrder string) bool {
	if sortOrder == "descending" {
		return a1.ID > a2.ID
	} else {
		return a1.ID < a2.ID
	}
}

func lessFunctionString(a1 *entity.Attendee, a2 *entity.Attendee, get func(a *entity.Attendee) string, sortOrder string) bool {
	if sortOrder == "descending" {
		return get(a1) > get(a2)
	} else {
		return get(a1) < get(a2)
	}
}

// --- admin info ---

func (r *InMemoryRepository) GetAdminInfoByAttendeeId(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error) {
	if ai, ok := r.adminInfo[attendeeId]; ok {
		// copy the info, so later modifications won't also modify it in the simulated db
		copiedAdminInfo := *ai
		return &copiedAdminInfo, nil
	} else {
		aiEmpty := entity.AdminInfo{}
		aiEmpty.ID = attendeeId
		return &aiEmpty, nil
	}
}

func (r *InMemoryRepository) WriteAdminInfo(ctx context.Context, ai *entity.AdminInfo) error {
	if ai.ID == 0 {
		return fmt.Errorf("cannot save admin info for attendee ID 0")
	}

	copiedAdminInfo := *ai
	r.adminInfo[ai.ID] = &copiedAdminInfo
	return nil
}

// --- status changes ---

func (r *InMemoryRepository) GetLatestStatusChangeByAttendeeId(ctx context.Context, attendeeId uint) (*entity.StatusChange, error) {
	scEmpty := entity.StatusChange{
		AttendeeId: attendeeId,
		Status:     status.New,
		Comments:   "",
	}
	if scList, ok := r.statusChanges[attendeeId]; ok {
		if len(scList) > 0 {
			sc := scList[len(scList)-1]
			return &sc, nil
		} else {
			return &scEmpty, nil
		}
	} else {
		return &scEmpty, nil
	}
}

func (r *InMemoryRepository) GetStatusChangesByAttendeeId(ctx context.Context, attendeeId uint) ([]entity.StatusChange, error) {
	if scList, ok := r.statusChanges[attendeeId]; ok {
		scListCopy := make([]entity.StatusChange, len(scList))
		for i := range scList {
			scListCopy[i] = scList[i]
		}
		return scListCopy, nil
	} else {
		return make([]entity.StatusChange, 0), nil
	}
}

func (r *InMemoryRepository) AddStatusChange(ctx context.Context, sc *entity.StatusChange) error {
	if scList, ok := r.statusChanges[sc.AttendeeId]; ok {
		scCopy := *sc
		scCopy.CreatedAt = time.Now()
		r.statusChanges[sc.AttendeeId] = append(scList, scCopy)
	} else {
		scCopy := *sc
		scCopy.CreatedAt = time.Now()
		r.statusChanges[sc.AttendeeId] = []entity.StatusChange{scCopy}
	}
	return nil
}

func (r *InMemoryRepository) FindByIdentity(ctx context.Context, identity string) ([]*entity.Attendee, error) {
	result := make([]*entity.Attendee, 0)
	for _, a := range r.attendees {
		if a.Identity == identity {
			copiedAttendee := *a
			result = append(result, &copiedAttendee)
		}
	}
	return result, nil
}

// --- bans ---

func (r *InMemoryRepository) GetAllBans(ctx context.Context) ([]*entity.Ban, error) {
	result := make([]*entity.Ban, 0)
	for _, b := range r.bans {
		copiedBan := *b
		result = append(result, &copiedBan)
	}
	sort.Slice(result, func(i int, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func (r *InMemoryRepository) GetBanById(ctx context.Context, id uint) (*entity.Ban, error) {
	b, ok := r.bans[id]
	if !ok {
		return &entity.Ban{}, fmt.Errorf("ban with id %d not found", id)
	}
	copiedBan := *b
	return &copiedBan, nil
}

func (r *InMemoryRepository) AddBan(ctx context.Context, b *entity.Ban) (uint, error) {
	newId := uint(atomic.AddUint32(&r.idSequence, 1))
	b.ID = newId

	// copy the ban, so later modifications won't also modify it in the simulated db
	copiedBan := *b
	r.bans[newId] = &copiedBan
	return newId, nil
}

func (r *InMemoryRepository) UpdateBan(ctx context.Context, b *entity.Ban) error {
	if _, ok := r.bans[b.ID]; ok {
		// copy the ban, so later modifications won't also modify it in the simulated db
		copiedBan := *b
		r.bans[b.ID] = &copiedBan
		return nil
	} else {
		return fmt.Errorf("cannot update ban %d - not present", b.ID)
	}
}

func (r *InMemoryRepository) DeleteBan(ctx context.Context, b *entity.Ban) error {
	if _, ok := r.bans[b.ID]; ok {
		delete(r.bans, b.ID)
		return nil
	} else {
		return fmt.Errorf("cannot delete ban %d - not present", b.ID)
	}
}

// --- additional info ---

func (r *InMemoryRepository) GetAllAdditionalInfoOrEmptyMap(ctx context.Context, attendeeId uint) map[string]*entity.AdditionalInfo {
	byAttendeeId, ok := r.addInfo[attendeeId]
	if !ok {
		byAttendeeId = make(map[string]*entity.AdditionalInfo)
	}
	return byAttendeeId
}

func (r *InMemoryRepository) GetAllAdditionalInfoForArea(ctx context.Context, area string) ([]*entity.AdditionalInfo, error) {
	result := make([]*entity.AdditionalInfo, 0)
	for attendeeId, areaMap := range r.addInfo {
		areaValue, ok := areaMap[area]
		if ok {
			entry := &entity.AdditionalInfo{
				AttendeeId: attendeeId,
				Area:       area,
				JsonValue:  areaValue.JsonValue,
			}
			result = append(result, entry)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) GetAdditionalInfoFor(ctx context.Context, attendeeId uint, area string) (*entity.AdditionalInfo, error) {
	byAttendeeId := r.GetAllAdditionalInfoOrEmptyMap(ctx, attendeeId)
	ai, ok := byAttendeeId[area]
	if !ok {
		// return a new entry suitable for saving
		return &entity.AdditionalInfo{
			AttendeeId: attendeeId,
			Area:       area,
		}, nil
	}
	copiedAi := *ai
	return &copiedAi, nil
}

func (r *InMemoryRepository) WriteAdditionalInfo(ctx context.Context, ad *entity.AdditionalInfo) error {
	if ad.Area == "" {
		return errors.New("invalid empty area for additional info")
	}

	byAttendeeId, ok := r.addInfo[ad.AttendeeId]
	if !ok {
		r.addInfo[ad.AttendeeId] = make(map[string]*entity.AdditionalInfo)
		byAttendeeId = r.addInfo[ad.AttendeeId]
	}

	origAd, ok := byAttendeeId[ad.Area]
	if !ok {
		// new
		ad.ID = uint(atomic.AddUint32(&r.idSequence, 1))
	} else {
		// existing
		if origAd.ID != ad.ID {
			return errors.New("unique constraint violated, please update the existing entry instead")
		}
	}
	copiedAi := *ad
	byAttendeeId[ad.Area] = &copiedAi
	return nil
}

// --- history ---

func (r *InMemoryRepository) RecordHistory(ctx context.Context, h *entity.History) error {
	newId := uint(atomic.AddUint32(&r.idSequence, 1))
	h.ID = newId
	r.history[newId] = h
	return nil
}

// only offered for testing, and only on the in memory db
func (r *InMemoryRepository) GetHistoryById(ctx context.Context, id uint) (*entity.History, error) {
	if h, ok := r.history[id]; ok {
		return h, nil
	} else {
		return &entity.History{}, fmt.Errorf("cannot get history entry %d - not present", id)
	}
}
