package acceptance

import (
	"testing"
)

func tstNoToken() string {
	return ""
}

const valid_JWT_is_not_staff = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZ2xvYmFsIjp7Im5hbWUiOiJKb2huIERvZSIsInJvbGVzIjpbXX0sImlhdCI6MTUxNjIzOTAyMn0.ove6_7BWQRe9HQyphwDdbiaAchgn9ynC4-2EYEXFeVTDADC4P3XYv5uLisYg4Mx8BZOnkWX-5L82pFO1mUZM147gLKMsYlc-iMKXy4sKZPzhQ_XKnBR-EBIf5x_ZD1wpva9ti7Yrvd0vDi8YSFdqqf7R4RA11hv9kg-_gg1uea6sK-Q_eEqoet7ocqGVLu-ghhkZdVLxu9tWJFPNueILWv8vW1Y_u9fDtfOhw7Ugf5ysI9RXiO-tXEHKN2HnFPCkwccnMFt4PJRzU1VoOldz0xzzZRb-j2tlbjLqcQkjMwLEoPQpC4Wbl8DgkaVdTi2aNyH7EbWMynlSOZIYK0AFvQ`
const valid_JWT_is_staff = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZ2xvYmFsIjp7Im5hbWUiOiJKb2huIERvZSIsInJvbGVzIjpbInN0YWZmIl19LCJpYXQiOjE1MTYyMzkwMjJ9.PNO4vV6V6iRg4-LcvJsRHyTSx7-6lDmqh6GrUWM4_OrhmmUWh2W4KF6sOfUco7sJ_I0PFOrnPGqREYAPuG1oAkHfitq5GdkYHCnJuHXXWo5d982GPs7zI-l9SxAgcUDdytesmSbq9Ktoad94OUL5bR8Uln0DPTlZvXDTAuCmAMW_89a4C-i71bsCYaFgL0RsJQ4yR4f3ez2M4hG4mNBjwaU4Ke77qdQIjx_9pP5ph37X8Z7twsC1yBH-Hev-293Naj3FZS8y63Zb6VGG3w8WW69eN_apoGRo26ZyaiDChAzOI-c1xkbMC5KYbnFQl5Ubdgk8sQgmp20RHHTV1R8Bcg`
const valid_JWT_is_admin = `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZ2xvYmFsIjp7Im5hbWUiOiJKb2huIERvZSIsInJvbGVzIjpbImFkbWluIl19LCJpYXQiOjE1MTYyMzkwMjJ9.sriAGCekreVU3nlQHc8Di7BqqI4Tut7tVNMWYa3kEpRi39Em5lOQ0b7w69idZEKT-MJfBGLVicnkw7Q4l8pUpJuHZMnja5YBIp7FDTg-KKbX__oOSSOnLhjaIGNFR_Xk_DanGrolQMKSYIfQs8MSgRO1bq-ZccCp1iJ4sdOOS4PenXj9h6xSe_lidGp8Wk47qwzRAFHYURaHFl_TCPMNDrYbM5MMIv8Lkye_duLxLo3zc9bnwWinhyD00p7ASwKgMc6vtWeTu_h000OOuviKoc2XKzOjUurqtm9Cird5rDAgAYtT_nTI_N4IzWFiRRPqX1IODe2zlqvKucv_FjzE8g`

func tstValidUserToken(t *testing.T, id string) string {
	return valid_JWT_is_not_staff
}

func tstValidAdminToken(t *testing.T) string {
	return valid_JWT_is_admin
}

func tstValidStaffToken(t *testing.T, id string) string {
	return valid_JWT_is_staff
}

func tstValidStaffOrEmptyToken(t *testing.T) string {
	return ""
}
