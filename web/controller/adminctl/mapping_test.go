package adminctl

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMapping(t *testing.T) {
	docs.Description("mapping an admin info dto back and forth should result in the same data")
	adminInfoDtoSource := tstCreateValidAdminInfo()
	adminInfoEntity := entity.AdminInfo{}
	mapDtoToAdminInfo(&adminInfoDtoSource, &adminInfoEntity)

	adminInfoDtoResult := admin.AdminInfoDto{}
	mapAdminInfoToDto(&adminInfoEntity, &adminInfoDtoResult)
	// id differences are ok because the field is only mapped one way, so overwrite with actual value
	adminInfoDtoSource.Id = adminInfoDtoResult.Id
	require.EqualValues(t, adminInfoDtoSource, adminInfoDtoResult, "unexpected difference after mapping back and forth")
}
