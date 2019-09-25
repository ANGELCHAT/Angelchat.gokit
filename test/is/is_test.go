package is_test

import (
	"testing"

	"github.com/livechat/gokit/test/is"
)

func TestMustAssert(t *testing.T) {
	is.True(t, 1 == 2, "not what?")
	is.True(t, 1 == 1, "")
}
