package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/authservice"
	"testing"
)

func tstNoToken() string {
	return ""
}

const valid_JWT_is_not_staff_sub1234567890 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJodHRwOi8vaWRlbnRpdHkubG9jYWxob3N0LyIsImp0aSI6IjQwNmJlM2U0LWY0ZTktNDdiNy1hYzVmLTA2YjkyNzQzMjg0OCIsIm5hbWUiOiJKb2huIERvZSIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjEyMzQ1Njc4OSJ9.XOy7LUJVsc7VBuintQDQ5asAbhmOEPzYNQwW0cxJhvlQMq77IBx1kUCCbg3_mstMopKQ85Njqhi5BksKpXuviRZE1BAzB5oQvIiB5IPyJrksm9Q5brJan37jclNc1rQN5wwAsGyY5alB4i9EeX4qo-ZWedtQPSdFTvUIOWf7-LpgWvc_xibQnPtbDwe1kkjbj6-fcubvkGI66yOylFGsg01jisYgWIIcV5N29KRffadJ2spk1tSCNvzTw-G4qcWHvBXQf2FUlOeKZSPV21-RwvHaTJYCyLCBt0CLDx847d44qaDBAxdntQI5KnhvEwthw-FvV0mPcgGA4fA-6l8v7A`
const valid_JWT_is_not_staff_sub101 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbInNvbWVncm91cCJdLCJpYXQiOjE1MTYyMzkwMjIsImlzcyI6Imh0dHA6Ly9pZGVudGl0eS5sb2NhbGhvc3QvIiwianRpIjoiNDA2YmUzZTQtZjRlOS00N2I3LWFjNWYtMDZiOTI3NDMyODQ4IiwibmFtZSI6IkpvaG4gRG9lIiwibm9uY2UiOiIzMGM4M2MxM2M5MTc5ODA0YWEwZjliMzkzNDI1OWQ3NSIsInJhdCI6MTY3NTExNzE3Nywic2lkIjoiZDdiOGZlN2EtMDc5YS00NTk2LThlNTMtYTYwZjg2YTA4YWM2Iiwic3ViIjoiMTAxIn0.ntHz3G7LLtHC3pJ1PoWJoG3mnzg96IIcP3LAV4V1CcKYMFoKVQfh7MiOdRXpiB-_j4QFE7O-za3mynwFqRbF3_Tw_Sp7Zsgk9OUPo2Mk3VBSl9yPIU4pmc8v7nrmaAVOQLyjglVG7NLRWLpx0oIG8SSN0d75PBI5iLyQ0H7Zu0npEu6xekHeAYAg9DHQxqZInzom72aLmHdtG7tOqOgN0XphiK7zmIqm5aCg7R9_J9s0UU0g16_Phxm3DaynufGCjEPE2YrSL7hY9UVT2nfrHO7MvVOEKMG3RaKUDjzqOkLawz9TcUJlUTBc1J-91zYbdXLHYT_2b4EW_qa1C-P3Ow`

const valid_JWT_is_staff_sub1234567890 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbInN0YWZmIl0sImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiaHR0cDovL2lkZW50aXR5LmxvY2FsaG9zdC8iLCJqdGkiOiI0MDZiZTNlNC1mNGU5LTQ3YjctYWM1Zi0wNmI5Mjc0MzI4NDgiLCJuYW1lIjoiSm9obiBTdGFmZiIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjEyMzQ1Njc4OTAifQ.GgzYXcFQf6q6xRxRgjJx2F8CCcAV-lYZ0ZS1Legv8_uEyZcyzX27hoPBwR1w4HcEEPK-QRQCKs4qj7Jyr0GRGNcN5ZFZZzo4LOUZsmU26Hc9YNzAzc9jin783yWrF5cH2QnUxpmH9TmQGG1yekDSNn3Mn2AB-0iyUAl_vHQ8REJPT_Cilhd5l0wxAy8Ht-Lal5pcz5LDJ9mFUTpBR1B614Aq6QBdShfeWXCYje7dGVvDRfFXxpQ4kRRog9dTkMAa0MyFJ3MgF2Uv53lmq7BDbcwSYed3beIHUqe7TLkImsG8jtpGKfcOadnW8qOGr7FI4AhJi_GKJzvnepz9jrVBNg`
const valid_JWT_is_staff_sub202 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbInN0YWZmIl0sImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiaHR0cDovL2lkZW50aXR5LmxvY2FsaG9zdC8iLCJqdGkiOiI0MDZiZTNlNC1mNGU5LTQ3YjctYWM1Zi0wNmI5Mjc0MzI4NDgiLCJuYW1lIjoiSm9obiBTdGFmZiIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjIwMiJ9.pM-jMGdjwNvHQMov8JQpRa79CBjHAUHpwElYRvUz_DvhkqcG4SrntVruAlJRS8D9CccflKeTjSEfOiS2l52p0qQ7ZeNPSRQ9nsr_EHDpB7UqcUszqVaBWtIhwkiwca_sxe-8L9A9hPSe_kH9dhDHVbhUsj9vp0HBIV89mtH3i3D6s3quRYtCe9puepkmyf5JC-2TSGoSiyURoFdqXSNRPEuv1FhlyVICqylfkINceCe8dt7lJCrhOc8R-11vA-SRsrBhdxBvYjT29hhQO3eHgJenPufjJPj6kYSWvN91U3KcsffoBmu-C1A7zBLu-zmWBHh4RkYWqbZpNr59TIpRSw`

const valid_JWT_is_admin_sub1234567890 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbImFkbWluIl0sImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiaHR0cDovL2lkZW50aXR5LmxvY2FsaG9zdC8iLCJqdGkiOiI0MDZiZTNlNC1mNGU5LTQ3YjctYWM1Zi0wNmI5Mjc0MzI4NDgiLCJuYW1lIjoiSm9obiBBZG1pbiIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjEyMzQ1Njc4OTAifQ.SpUNhK20KP0HDGAgjhejnoY-iV4hJm58VFDNDtE5H2vonEywgWg50emPDgJl7K3DR8a40QfJiMpJuGYoHf7S7WZeUOhfyfsL0gY2p5E1X_bdL1viUHJJwoB6mUNrSmieYFURkkphBbnRdUEgNOMuq-2d8i3Zp_YxNfEArqkpxJDqVeWJS8koeJTH2GVcWOjryXmo3SbextZwmkwL1HryDgQkd2iSgtYUSih01-wnzPPvWyEpfw58w7uPIej5uAK6kdUNLZBYtAB-9XbbcDVuDSPmBrMP0qZWiYtVjq5oeEbOX7ucNB3qT_UR2Mk2oG8kD_d2vcLdBq8CmrAjQ6lHnQ`

func tstValidUserToken(t *testing.T, id uint) string {
	if id == 101 {
		return valid_JWT_is_not_staff_sub101
	} else {
		return valid_JWT_is_not_staff_sub1234567890
	}
}

func tstValidAdminToken(t *testing.T) string {
	return valid_JWT_is_admin_sub1234567890
}

func tstValidStaffToken(t *testing.T, id uint) string {
	if id == 202 {
		return valid_JWT_is_staff_sub202
	} else {
		return valid_JWT_is_staff_sub1234567890
	}
}

func tstValidStaffOrEmptyToken(t *testing.T) string {
	return ""
}

const valid_Api_Token_Matches_Test_Configuration_Files = "api-token-for-testing-must-be-pretty-long"

func tstValidApiToken() string {
	return valid_Api_Token_Matches_Test_Configuration_Files
}

func tstInvalidApiToken() string {
	return "wrong_api_token"
}

func tstSetupAuthMockResponses() {
	// we pretend the id token is also an access token, but with a prefix
	authMock.SetupResponse(valid_JWT_is_not_staff_sub1234567890, "access"+valid_JWT_is_not_staff_sub1234567890, authservice.UserInfoResponse{
		Subject:       "1234567890",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse(valid_JWT_is_not_staff_sub101, "access"+valid_JWT_is_not_staff_sub101, authservice.UserInfoResponse{
		Subject:       "101",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse(valid_JWT_is_staff_sub1234567890, "access"+valid_JWT_is_staff_sub1234567890, authservice.UserInfoResponse{
		Subject:       "1234567890",
		Name:          "John Staff",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"staff"},
	})
	authMock.SetupResponse(valid_JWT_is_staff_sub202, "access"+valid_JWT_is_staff_sub202, authservice.UserInfoResponse{
		Subject:       "202",
		Name:          "John Staff",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"staff"},
	})
	authMock.SetupResponse(valid_JWT_is_admin_sub1234567890, "access"+valid_JWT_is_admin_sub1234567890, authservice.UserInfoResponse{
		Subject:       "1234567890",
		Name:          "John Admin",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"admin"},
	})

	// also accept auth with just the access token
	authMock.SetupResponse("", "access"+valid_JWT_is_not_staff_sub1234567890, authservice.UserInfoResponse{
		Subject:       "1234567890",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse("", "access"+valid_JWT_is_not_staff_sub101, authservice.UserInfoResponse{
		Subject:       "101",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse("", "access"+valid_JWT_is_staff_sub1234567890, authservice.UserInfoResponse{
		Subject:       "1234567890",
		Name:          "John Staff",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"staff"},
	})
	authMock.SetupResponse("", "access"+valid_JWT_is_staff_sub202, authservice.UserInfoResponse{
		Subject:       "202",
		Name:          "John Staff",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"staff"},
	})
	authMock.SetupResponse("", "access"+valid_JWT_is_admin_sub1234567890, authservice.UserInfoResponse{
		Subject:       "1234567890",
		Name:          "John Admin",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"admin"},
	})
}
