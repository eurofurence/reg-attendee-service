package banctl

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/bans"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
)

func mapDtoToBan(dto *bans.BanRule, b *entity.Ban) {
	// do not map id - instead load by ID from db, or you'll introduce errors
	b.Reason = dto.Reason
	b.NamePattern = dto.NamePattern
	b.NicknamePattern = dto.NicknamePattern
	b.EmailPattern = dto.EmailPattern
}

func mapBanToDto(b *entity.Ban, dto *bans.BanRule) {
	dto.Id = b.ID
	dto.Reason = b.Reason
	dto.NamePattern = b.NamePattern
	dto.NicknamePattern = b.NicknamePattern
	dto.EmailPattern = b.EmailPattern
}
