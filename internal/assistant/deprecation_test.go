package assistant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeprecationInfo_ZeroValue(t *testing.T) {
	var d DeprecationInfo
	assert.False(t, d.Deprecated(), "zero value should not be deprecated")
	assert.Equal(t, "", d.DeprecatedRemovedIn(), "zero value should return empty string for RemovedIn")
}

func TestDeprecationInfo_Deprecated(t *testing.T) {
	d := DeprecationInfo{
		IsDeprecated: true,
		RemovedIn:    "v2.0.0",
	}
	assert.True(t, d.Deprecated(), "should be deprecated")
	assert.Equal(t, "v2.0.0", d.DeprecatedRemovedIn(), "should return the RemovedIn version")
}

func TestDeprecationInfo_DeprecatedWithoutVersion(t *testing.T) {
	d := DeprecationInfo{
		IsDeprecated: true,
	}
	assert.True(t, d.Deprecated(), "should be deprecated")
	assert.Equal(t, "", d.DeprecatedRemovedIn(), "should return empty string when RemovedIn not set")
}
