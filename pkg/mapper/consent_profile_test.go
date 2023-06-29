package mapper

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

type TestCase struct {
	name     string
	profile  *ConsentProfile
	template string
	expected *ConsentPolicy
}

func TestGetPolicy(t *testing.T) {

	mii := NewConsentProfile(MiiProfile)
	cases := []TestCase{
		{"consentTemplate", mii, "Patienteneinwilligung MII|1.6.d", mii.ConsentPolicy},
		{"withdrawalTemplate", mii, "Teilwiderruf (kompatibel zu Patienteneinwilligung MII 1.6d)|2.0.a", mii.ConsentPolicy},
		{"completeWithdrawalTemplate", mii, "Vollständiger Widerruf (kompatibel zu Patienteneinwilligung MII 1.6d)", mii.ConsentPolicy},
		{"garbledWithdrawalTemplate", mii, "Vollst�ndiger Widerruf (kompatibel zu Patienteneinwilligung MII 1.6d)", mii.ConsentPolicy},
		{"invalidWithdrawalTemplate", mii, "Patienteneinwilligung Projekt 42|1.0", nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			runGetPolicy(t, c)
		})
	}
}

func runGetPolicy(t *testing.T, data TestCase) {
	actual := data.profile.GetPolicy(data.template)

	assert.Equal(t, actual, data.expected)
}
