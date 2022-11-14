package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/bans"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
)

// ------------------------------------------
// acceptance tests for the ban rule resource
// ------------------------------------------

// --- create new ban rule ---

func TestCreateNewBanRule_Success(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new ban rule with valid data")
	banSent := tstBuildValidBanRule("ban11-")
	response := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), token)

	docs.Then("then the ban is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/bans\\/[1-9][0-9]*$", response.location, "invalid location header in response")

	docs.Then("and it can be read again")
	banReadAgain := tstReadBan(t, response.location)
	// difference in id is ok, so copy it over to expected
	banSent.Id = banReadAgain.Id
	require.EqualValues(t, banSent, banReadAgain, "ban rule data read did not match sent data")
}

func TestCreateNewBanRule_Unauthenticated(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an anonymous user")

	docs.When("when they attempt to create a new ban rule with valid data")
	banSent := tstBuildValidBanRule("ban12-")
	response := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and the ban has not been added")
	tstRequireBanDbSize(t, 0)
}

func TestCreateNewBanRule_Unauthorized_User(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given a regular user")
	token := tstValidUserToken(t, "22")

	docs.When("when they attempt to create a new ban rule with valid data")
	banSent := tstBuildValidBanRule("ban13-")
	response := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and the ban has not been added")
	tstRequireBanDbSize(t, 0)
}

func TestCreateNewBanRule_Unauthorized_Staff(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given a staffer")
	token := tstValidStaffToken(t, "42")

	docs.When("when they attempt to create a new ban rule with valid data")
	banSent := tstBuildValidBanRule("ban14-")
	response := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and the ban has not been added")
	tstRequireBanDbSize(t, 0)
}

func TestCreateNewBanRule_InvalidJson(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new ban rule but send syntactically invalid json")
	response := tstPerformPost("/api/rest/v1/bans", "{{{!!! ban15- &", token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "ban.parse.error", "")
}

func TestCreateNewBanRule_InvalidValues(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new ban rule with invalid data")
	banSent := tstBuildValidBanRule("ban16-")
	banSent.EmailPattern = "(unmatched|parens\\)"
	banSent.Id = "57"
	banSent.NamePattern = "****"
	banSent.NicknamePattern = "unclosed[group"
	banSent.Reason = ""
	response := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "ban.data.invalid", url.Values{
		"email_pattern":    []string{"email_pattern field must be empty or contain a valid regular expression: error parsing regexp: missing closing ): `(unmatched|parens\\)`"},
		"id":               []string{"id field must be empty or correctly assigned for incoming requests"},
		"name_pattern":     []string{"name_pattern field must be empty or contain a valid regular expression: error parsing regexp: missing argument to repetition operator: `*`"},
		"nickname_pattern": []string{"nickname_pattern field must be empty or contain a valid regular expression: error parsing regexp: missing closing ]: `[group`"},
		"reason":           []string{"reason field must be at least 1 and at most 255 characters long"},
	})

	docs.Then("and the ban has not been added")
	tstRequireBanDbSize(t, 0)
}

func TestCreateNewBanRule_Duplicate(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.Given("given an existing ban rule")
	banSent := tstBuildValidBanRule("ban17-")
	response1 := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), token)
	require.Equal(t, http.StatusCreated, response1.status, "failed to create test prerequisites")

	docs.When("when they attempt to create a duplicate of the existing new ban rule")
	response := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusConflict, "ban.data.duplicate", url.Values{
		"ban": []string{"there is already another ban rule with the same patterns"},
	})

	docs.Then("and the duplicate ban rule has not been added")
	tstRequireBanDbSize(t, 1)
}

func TestCreateNewBanRule_NoDuplicate(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.Given("given an existing ban rule")
	banSent := tstBuildValidBanRule("ban18-")
	response1 := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), token)
	require.Equal(t, http.StatusCreated, response1.status, "failed to create test prerequisites")

	docs.When("when they create a duplicate of the existing new ban rule, but one of the patterns differs")
	banSent.NamePattern = "different.*pattern"
	response := tstPerformPost("/api/rest/v1/bans", tstRenderJson(banSent), token)

	docs.Then("then the ban is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")

	docs.Then("and the second ban rule has been added")
	tstRequireBanDbSize(t, 2)
}

// --- read all existing ban rules ---

func TestReadBanRules_Success(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	ban1, _, ban2, _ := tstCreatePreexistingBans(t, "ban21-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to list the ban rules")
	response := tstPerformGet("/api/rest/v1/bans", token)

	docs.Then("then the request is successful and the correct information is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	dto := bans.BanRuleList{}
	tstParseJson(response.body, &dto)
	require.NotNil(t, dto.Bans)
	require.Equal(t, 2, len(dto.Bans))
	require.EqualValues(t, ban1, dto.Bans[0])
	require.EqualValues(t, ban2, dto.Bans[1])
}

func TestReadBanRules_Unauthenticated(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, _, _ = tstCreatePreexistingBans(t, "ban22-")

	docs.Given("given an anonymous user")

	docs.When("when they attempt to list the ban rules")
	response := tstPerformGet("/api/rest/v1/bans", tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestReadBanRules_Unauthorized_User(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, _, _ = tstCreatePreexistingBans(t, "ban23-")

	docs.Given("given a regular user")
	token := tstValidUserToken(t, "22")

	docs.When("when they attempt to list the ban rules")
	response := tstPerformGet("/api/rest/v1/bans", token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestReadBanRules_Unauthorized_Staff(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, _, _ = tstCreatePreexistingBans(t, "ban24-")

	docs.Given("given a staffer")
	token := tstValidStaffToken(t, "42")

	docs.When("when they attempt to list the ban rules")
	response := tstPerformGet("/api/rest/v1/bans", token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

// --- read a single ban rule ---

func TestReadBanRule_Success(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, ban2, loc2 := tstCreatePreexistingBans(t, "ban31-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to read an existing ban rule")
	response := tstPerformGet(loc2, token)

	docs.Then("then the request is successful and the correct information is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	dto := bans.BanRule{}
	tstParseJson(response.body, &dto)
	require.EqualValues(t, ban2, dto)
}

func TestReadBanRule_Unauthenticated(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, _, loc2 := tstCreatePreexistingBans(t, "ban32-")

	docs.Given("given an anonymous user")

	docs.When("when they attempt to read an existing ban rule")
	response := tstPerformGet(loc2, tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestReadBanRule_Unauthorized_User(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, _, loc2 := tstCreatePreexistingBans(t, "ban33-")

	docs.Given("given a regular user")
	token := tstValidUserToken(t, "22")

	docs.When("when they attempt to read an existing ban rule")
	response := tstPerformGet(loc2, token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestReadBanRule_Unauthorized_Staff(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, _, loc2 := tstCreatePreexistingBans(t, "ban34-")

	docs.Given("given a staffer")
	token := tstValidStaffToken(t, "42")

	docs.When("when they attempt to read an existing ban rule")
	response := tstPerformGet(loc2, token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestReadBanRule_InvalidId(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to read a ban rule, but supply an invalid id")
	response := tstPerformGet("/api/rest/v1/bans/kitten", token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "ban.id.invalid", url.Values{})
}

func TestReadBanRule_NotFound(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to read a ban rule, but supply an id that does not exist")
	response := tstPerformGet("/api/rest/v1/bans/42", token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "ban.id.notfound", url.Values{})
}

// --- update a ban rule

func TestUpdateBanRule_Success(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, ban2, loc2 := tstCreatePreexistingBans(t, "ban41-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to update an existing ban rule")
	ban2.Reason = "made a change to reason"
	response := tstPerformPut(loc2, tstRenderJson(ban2), token)

	docs.Then("then the request is successful")
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")

	docs.Then("and the changed ban can be read again")
	ban2reread := tstReadBan(t, loc2)
	require.EqualValues(t, ban2, ban2reread, "ban was not updated correctly")
}

func TestUpdateBanRule_Unauthenticated(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, ban2, loc2 := tstCreatePreexistingBans(t, "ban42-")

	docs.Given("given an anonymous user")

	docs.When("when they attempt to update an existing ban rule")
	banUpdated := ban2
	banUpdated.Reason = "made a change to reason"
	response := tstPerformPut(loc2, tstRenderJson(banUpdated), tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and no change has been made to the ban rule")
	ban2reread := tstReadBan(t, loc2)
	require.EqualValues(t, ban2, ban2reread, "ban was unexpectedly changed")
}

func TestUpdateBanRule_Unauthorized_User(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, ban2, loc2 := tstCreatePreexistingBans(t, "ban43-")

	docs.Given("given a regular user")
	token := tstValidUserToken(t, "22")

	docs.When("when they attempt to update an existing ban rule")
	banUpdated := ban2
	banUpdated.Reason = "made a change to reason"
	response := tstPerformPut(loc2, tstRenderJson(banUpdated), token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and no change has been made to the ban rule")
	ban2reread := tstReadBan(t, loc2)
	require.EqualValues(t, ban2, ban2reread, "ban was unexpectedly changed")
}

func TestUpdateBanRule_Unauthorized_Staff(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, ban2, loc2 := tstCreatePreexistingBans(t, "ban44-")

	docs.Given("given a staffer")
	token := tstValidStaffToken(t, "42")

	docs.When("when they attempt to update an existing ban rule")
	banUpdated := ban2
	banUpdated.Reason = "made a change to reason"
	response := tstPerformPut(loc2, tstRenderJson(banUpdated), token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and no change has been made to the ban rule")
	ban2reread := tstReadBan(t, loc2)
	require.EqualValues(t, ban2, ban2reread, "ban was unexpectedly changed")
}

func TestUpdateBanRule_InvalidId(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to update a ban rule, but supply an invalid id")
	ban := tstBuildValidBanRule("ban45-")
	response := tstPerformPut("/api/rest/v1/bans/puppy", tstRenderJson(ban), token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "ban.id.invalid", url.Values{})
}

func TestUpdateBanRule_NotFound(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to update a ban rule, but supply an id that does not exist")
	ban := tstBuildValidBanRule("ban46-")
	ban.Id = "46"
	response := tstPerformPut("/api/rest/v1/bans/46", tstRenderJson(ban), token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "ban.id.notfound", url.Values{})
}

func TestUpdateBanRule_InvalidJson(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, ban2, loc2 := tstCreatePreexistingBans(t, "ban46-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to update an existing ban rule but send syntactically invalid json")
	response := tstPerformPut(loc2, "{{{!!! ban47- &", token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "ban.parse.error", "")

	docs.Then("and no change has been made to the ban rule")
	ban2reread := tstReadBan(t, loc2)
	require.EqualValues(t, ban2, ban2reread, "ban was unexpectedly changed")
}

func TestUpdateBanRule_InvalidValues(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	_, _, ban2, loc2 := tstCreatePreexistingBans(t, "ban47-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to update an existing ban rule but send a body with invalid values")
	ban2updated := ban2
	ban2updated.NicknamePattern = "])*"
	response := tstPerformPut(loc2, tstRenderJson(ban2updated), token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "ban.data.invalid", url.Values{
		"nickname_pattern": []string{"nickname_pattern field must be empty or contain a valid regular expression: error parsing regexp: unexpected ): `])*`"},
	})

	docs.Then("and no change has been made to the ban rule")
	ban2reread := tstReadBan(t, loc2)
	require.EqualValues(t, ban2, ban2reread, "ban was unexpectedly changed")
}

func TestUpdateBanRule_Duplicate(t *testing.T) {
	tstSetup(tstConfigFileIrrelevant())
	defer tstShutdown()

	docs.Given("given existing ban rules")
	ban1, _, ban2, loc2 := tstCreatePreexistingBans(t, "ban47-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to update an existing ban rule in a way that would cause a duplicate")
	ban2updated := ban2
	ban2updated.NamePattern = ban1.NamePattern
	ban2updated.NicknamePattern = ban1.NicknamePattern
	ban2updated.EmailPattern = ban1.EmailPattern
	response := tstPerformPut(loc2, tstRenderJson(ban2updated), token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusConflict, "ban.data.duplicate", url.Values{
		"ban": []string{"there is already another ban rule with the same patterns"},
	})

	docs.Then("and no change has been made to the ban rule")
	ban2reread := tstReadBan(t, loc2)
	require.EqualValues(t, ban2, ban2reread, "ban was unexpectedly changed")
}

// helper functions

func tstReadBan(t *testing.T, location string) bans.BanRule {
	readAgainResponse := tstPerformGet(location, tstValidAdminToken(t))
	banReadAgain := bans.BanRule{}
	tstParseJson(readAgainResponse.body, &banReadAgain)
	return banReadAgain
}

func tstRequireBanDbSize(t *testing.T, expected int) {
	response := tstPerformGet("/api/rest/v1/bans", tstValidAdminToken(t))
	dto := bans.BanRuleList{}
	tstParseJson(response.body, &dto)
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	require.NotNil(t, dto.Bans)
	require.Equal(t, expected, len(dto.Bans), "found unexpected ban list length")
}

func tstCreatePreexistingBans(t *testing.T, testcase string) (bans.BanRule, string, bans.BanRule, string) {
	token := tstValidAdminToken(t)
	ban1 := tstBuildValidBanRule(testcase + "a-")
	response1 := tstPerformPost("/api/rest/v1/bans", tstRenderJson(ban1), token)
	require.Equal(t, http.StatusCreated, response1.status, "unexpected http response status")
	ban2 := tstBuildValidBanRule(testcase + "b-")
	response2 := tstPerformPost("/api/rest/v1/bans", tstRenderJson(ban2), token)
	require.Equal(t, http.StatusCreated, response2.status, "unexpected http response status")
	ban1.Id = tstIdFromLoc(response1.location)
	ban2.Id = tstIdFromLoc(response2.location)
	return ban1, response1.location, ban2, response2.location
}

func tstIdFromLoc(loc string) string {
	sections := strings.Split(loc, "/")
	return sections[len(sections)-1]
}
