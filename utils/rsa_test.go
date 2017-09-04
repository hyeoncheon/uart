package utils_test

import (
	"testing"

	"github.com/hyeoncheon/uart/utils"
	"github.com/stretchr/testify/require"
)

func Test_A_GenRSAKeyPair(t *testing.T) {
	r := require.New(t)
	_, _, err := utils.GenRSAKeyPair()
	r.NoError(err)
}
