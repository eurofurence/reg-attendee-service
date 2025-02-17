package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/authservice"
	"testing"
)

func tstNoToken() string {
	return ""
}

const valid_JWT_is_not_staff_sub1234567890 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJodHRwOi8vaWRlbnRpdHkubG9jYWxob3N0LyIsImp0aSI6IjQwNmJlM2U0LWY0ZTktNDdiNy1hYzVmLTA2YjkyNzQzMjg0OCIsIm5hbWUiOiJKb2huIERvZSIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjEyMzQ1Njc4OTAifQ.dOE4B-UkCZMpGwEERTD34AvFFM_VJSAMo-N1n3JrusVfcazfq8MBQ0LEr32stUrxAQAhPAaLHr2IlsUxYGhJ-OE5-oDI2n3-7_ixpqMLZKITgEd-RWkF89KSINJ8o53o_IFC8IgdYCIlC60II23TX7gkUJIAEKbRDIK08PgFep7c3LZygj-HZ54X_Q4nIJ5HtTD88XuedQSP9zd79R62dypGvpc38otv4fkw-u_lphDIVO8AzT6ZmscPnBu-oDRJqUlEpfvpcrw84kULpwnw5j1N-D54v4MJ36Y3LYA1_WCyRLECfacbX893CV7Khm4vlSg6fYCC-_PHo-2NMDHHDA`
const valid_JWT_is_not_staff_sub101 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbInNvbWVncm91cCJdLCJpYXQiOjE1MTYyMzkwMjIsImlzcyI6Imh0dHA6Ly9pZGVudGl0eS5sb2NhbGhvc3QvIiwianRpIjoiNDA2YmUzZTQtZjRlOS00N2I3LWFjNWYtMDZiOTI3NDMyODQ4IiwibmFtZSI6IkpvaG4gRG9lIiwibm9uY2UiOiIzMGM4M2MxM2M5MTc5ODA0YWEwZjliMzkzNDI1OWQ3NSIsInJhdCI6MTY3NTExNzE3Nywic2lkIjoiZDdiOGZlN2EtMDc5YS00NTk2LThlNTMtYTYwZjg2YTA4YWM2Iiwic3ViIjoiMTAxIn0.ntHz3G7LLtHC3pJ1PoWJoG3mnzg96IIcP3LAV4V1CcKYMFoKVQfh7MiOdRXpiB-_j4QFE7O-za3mynwFqRbF3_Tw_Sp7Zsgk9OUPo2Mk3VBSl9yPIU4pmc8v7nrmaAVOQLyjglVG7NLRWLpx0oIG8SSN0d75PBI5iLyQ0H7Zu0npEu6xekHeAYAg9DHQxqZInzom72aLmHdtG7tOqOgN0XphiK7zmIqm5aCg7R9_J9s0UU0g16_Phxm3DaynufGCjEPE2YrSL7hY9UVT2nfrHO7MvVOEKMG3RaKUDjzqOkLawz9TcUJlUTBc1J-91zYbdXLHYT_2b4EW_qa1C-P3Ow`
const valid_JWT_ev_sub102 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbImV2IiwiZnVyIl0sImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiaHR0cDovL2lkZW50aXR5LmxvY2FsaG9zdC8iLCJqdGkiOiI0MDZiZTNlNC1mNGU5LTQ3YjctYWM1Zi0wNmI5Mjc0MzI4NDgiLCJuYW1lIjoiSm9obiBEb2UiLCJub25jZSI6IjMwYzgzYzEzYzkxNzk4MDRhYTBmOWIzOTM0MjU5ZDc1IiwicmF0IjoxNjc1MTE3MTc3LCJzaWQiOiJkN2I4ZmU3YS0wNzlhLTQ1OTYtOGU1My1hNjBmODZhMDhhYzYiLCJzdWIiOiIxMDIifQ.qzHiYNkcr8Hkqpe86F_C849Z06TS1ZxkFYsiqvFFS__mVkbSS9jbUhCJNfckCc0dZleTfN8L1w7RK0fD1PQR3hsF-Wy4sZE9-ZzW7P1sNmYkmY68w4avpAMs7Fn3_o9Ros25oOqcEbu0d4M43GYDX8dwA729Jtle8N46LjJXhuYG6wz_K59qVd8kTMbUgm5GapWdrQs4Qlswnf_K1G5HXhAi7mrrMZOGejDODeofHPGukY1TZfMfMEUgJmlIn2nn6hu8fyyvpIDgaQpg1LKKw5JYzVi_EAjqz0xzXzvsJ1Tacj2aoXFDCxOawG-6-ID2Q4uPAJvZ9GTdmmePsJuhxw`
const valid_JWT_is_not_staff_sub101_unverified = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJleHAiOjIwNzUxMjA4MTYsImdyb3VwcyI6WyJzb21lZ3JvdXAiXSwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJodHRwOi8vaWRlbnRpdHkubG9jYWxob3N0LyIsImp0aSI6IjQwNmJlM2U0LWY0ZTktNDdiNy1hYzVmLTA2YjkyNzQzMjg0OCIsIm5hbWUiOiJKb2huIERvZSIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjEwMSJ9.QewwmuCatUYhcJPk_JZPeOqJOmh0XlbT9CKWPmjXT-ODX-oWZ2Dop3-J2xsMRSbMn23m1mXy8SXcUjIuFFzMcZCZY6O2-HD9igskn6e8yg8WBi2QnP-sOrWfvaLfnVORYwVxyO3o9eeWPhPjDaFVGvg7rzho_IVIXg0LqluN2ID3RcBc5JuzDGwm0YpuC9gJr1I5rDLADbXF3pLVDTGWFXrlln_1vbzhnPvKAJNPFhKwtuIEmKuLC9OgzW4bIjbPHU_A4dCfa7aAZ4D2RId7rBUOyVKIXQR0_K7UwIjx-oJlDyQsj0OSzgGsj6FUMJSZMI8lXOdH1i1haWc7ekbZqg`

const valid_JWT_is_staff_sub1234567890 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbInN0YWZmIl0sImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiaHR0cDovL2lkZW50aXR5LmxvY2FsaG9zdC8iLCJqdGkiOiI0MDZiZTNlNC1mNGU5LTQ3YjctYWM1Zi0wNmI5Mjc0MzI4NDgiLCJuYW1lIjoiSm9obiBTdGFmZiIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjEyMzQ1Njc4OTAifQ.GgzYXcFQf6q6xRxRgjJx2F8CCcAV-lYZ0ZS1Legv8_uEyZcyzX27hoPBwR1w4HcEEPK-QRQCKs4qj7Jyr0GRGNcN5ZFZZzo4LOUZsmU26Hc9YNzAzc9jin783yWrF5cH2QnUxpmH9TmQGG1yekDSNn3Mn2AB-0iyUAl_vHQ8REJPT_Cilhd5l0wxAy8Ht-Lal5pcz5LDJ9mFUTpBR1B614Aq6QBdShfeWXCYje7dGVvDRfFXxpQ4kRRog9dTkMAa0MyFJ3MgF2Uv53lmq7BDbcwSYed3beIHUqe7TLkImsG8jtpGKfcOadnW8qOGr7FI4AhJi_GKJzvnepz9jrVBNg`
const valid_JWT_is_staff_sub202 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbInN0YWZmIl0sImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiaHR0cDovL2lkZW50aXR5LmxvY2FsaG9zdC8iLCJqdGkiOiI0MDZiZTNlNC1mNGU5LTQ3YjctYWM1Zi0wNmI5Mjc0MzI4NDgiLCJuYW1lIjoiSm9obiBTdGFmZiIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjIwMiJ9.pM-jMGdjwNvHQMov8JQpRa79CBjHAUHpwElYRvUz_DvhkqcG4SrntVruAlJRS8D9CccflKeTjSEfOiS2l52p0qQ7ZeNPSRQ9nsr_EHDpB7UqcUszqVaBWtIhwkiwca_sxe-8L9A9hPSe_kH9dhDHVbhUsj9vp0HBIV89mtH3i3D6s3quRYtCe9puepkmyf5JC-2TSGoSiyURoFdqXSNRPEuv1FhlyVICqylfkINceCe8dt7lJCrhOc8R-11vA-SRsrBhdxBvYjT29hhQO3eHgJenPufjJPj6kYSWvN91U3KcsffoBmu-C1A7zBLu-zmWBHh4RkYWqbZpNr59TIpRSw`

const valid_JWT_is_admin_sub1234567890 = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdF9oYXNoIjoidDdWYkV5NVQ3STdYSlh3VHZ4S3hLdyIsImF1ZCI6WyIxNGQ5ZjM3YS0xZWVjLTQ3YzktYTk0OS01ZjFlYmRmOWM4ZTUiXSwiYXV0aF90aW1lIjoxNTE2MjM5MDIyLCJlbWFpbCI6ImpzcXVpcnJlbF9naXRodWJfOWE2ZEBwYWNrZXRsb3NzLmRlIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MjA3NTEyMDgxNiwiZ3JvdXBzIjpbImFkbWluIl0sImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiaHR0cDovL2lkZW50aXR5LmxvY2FsaG9zdC8iLCJqdGkiOiI0MDZiZTNlNC1mNGU5LTQ3YjctYWM1Zi0wNmI5Mjc0MzI4NDgiLCJuYW1lIjoiSm9obiBBZG1pbiIsIm5vbmNlIjoiMzBjODNjMTNjOTE3OTgwNGFhMGY5YjM5MzQyNTlkNzUiLCJyYXQiOjE2NzUxMTcxNzcsInNpZCI6ImQ3YjhmZTdhLTA3OWEtNDU5Ni04ZTUzLWE2MGY4NmEwOGFjNiIsInN1YiI6IjEyMzQ1Njc4OTAifQ.SpUNhK20KP0HDGAgjhejnoY-iV4hJm58VFDNDtE5H2vonEywgWg50emPDgJl7K3DR8a40QfJiMpJuGYoHf7S7WZeUOhfyfsL0gY2p5E1X_bdL1viUHJJwoB6mUNrSmieYFURkkphBbnRdUEgNOMuq-2d8i3Zp_YxNfEArqkpxJDqVeWJS8koeJTH2GVcWOjryXmo3SbextZwmkwL1HryDgQkd2iSgtYUSih01-wnzPPvWyEpfw58w7uPIej5uAK6kdUNLZBYtAB-9XbbcDVuDSPmBrMP0qZWiYtVjq5oeEbOX7ucNB3qT_UR2Mk2oG8kD_d2vcLdBq8CmrAjQ6lHnQ`

func tstValidUserToken(t *testing.T, id uint) string {
	if id == 101 {
		return valid_JWT_is_not_staff_sub101
	} else if id == 102 {
		return valid_JWT_ev_sub102
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
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "1234567890",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse(valid_JWT_is_not_staff_sub101, "access"+valid_JWT_is_not_staff_sub101, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "101",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse(valid_JWT_ev_sub102, "access"+valid_JWT_ev_sub102, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "102",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse(valid_JWT_is_staff_sub1234567890, "access"+valid_JWT_is_staff_sub1234567890, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "1234567890",
		Name:          "John Staff",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"staff"},
	})
	authMock.SetupResponse(valid_JWT_is_staff_sub202, "access"+valid_JWT_is_staff_sub202, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "202",
		Name:          "John Staff",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"staff"},
	})
	authMock.SetupResponse(valid_JWT_is_admin_sub1234567890, "access"+valid_JWT_is_admin_sub1234567890, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "1234567890",
		Name:          "John Admin",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"admin"},
	})

	// also accept auth with just the access token
	authMock.SetupResponse("", "access"+valid_JWT_is_not_staff_sub1234567890, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "1234567890",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse("", "access"+valid_JWT_is_not_staff_sub101, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "101",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse("", "access"+valid_JWT_ev_sub102, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "102",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})
	authMock.SetupResponse("", "access"+valid_JWT_is_staff_sub1234567890, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "1234567890",
		Name:          "John Staff",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"staff"},
	})
	authMock.SetupResponse("", "access"+valid_JWT_is_staff_sub202, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "202",
		Name:          "John Staff",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"staff"},
	})
	authMock.SetupResponse("", "access"+valid_JWT_is_admin_sub1234567890, authservice.UserInfoResponse{
		Audiences:     []string{"14d9f37a-1eec-47c9-a949-5f1ebdf9c8e5"},
		Subject:       "1234567890",
		Name:          "John Admin",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
		Groups:        []string{"admin"},
	})

	// setup auth response with a different audience
	authMock.SetupResponse("", "access_other_audience_101", authservice.UserInfoResponse{
		Audiences:     []string{"meow"},
		Subject:       "101",
		Name:          "John Doe",
		Email:         "jsquirrel_github_9a6d@packetloss.de",
		EmailVerified: true,
	})

}
