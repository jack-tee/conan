package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FormatPollIntervalWithMs(t *testing.T) {
	res := FormatPollInterval(954)
	assert.Equal(t, "954ms", res)
}

func Test_FormatPollIntervalWithSeconds(t *testing.T) {
	res := FormatPollInterval(5001)
	assert.Equal(t, "5s", res)
}

func Test_FormatPollIntervalWithMinutes(t *testing.T) {
	res := FormatPollInterval(130000)
	assert.Equal(t, "2m 10s", res)
}

func Test_FormatPollIntervalWithHours(t *testing.T) {
	res := FormatPollInterval(3600000)
	assert.Equal(t, "1h 0m", res)
}
