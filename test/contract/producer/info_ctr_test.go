package producer

import (
	"fmt"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

// see setup_ctr_test.go for http test server and service mock

// tests are run in the directory they are located in
// normally we would use a web server to which we publish the contracts, but this is fine for this example
const demoClientPactDir = "../../../../rexis-go-democlient/test/contract/consumer/pacts/"

// contract test provider side (a very contrived example)

func TestProvider(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "DemoClient",
		Provider: "AttendeeService",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Start provider API in the background
	// this is done during test startup using httptest package

	// Verify the Provider using the locally saved Pact Files
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: ts.URL,
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/democlient-attendeeservice.json", demoClientPactDir))},
		StateHandlers: 			types.StateHandlers{
			// Setup any state required by the test
			// example that we are not really using in this test
			"Attendee 1 exists": func() error {
				// set up service mock responses here if needed
				return nil
			},
		},
	})
	require.Nil(t, err, "unexpected error during verification")
	// now use the service mock to assert further expectations of what calls to the mock service should have
	// occurred during the verification.
}
