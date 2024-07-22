package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSetSignatureBanana(t *testing.T) {
	sut := SignedSequenceBanana{}
	signature := []byte{1, 2, 3}
	sut.SetSignature(signature)
	assert.Equal(t, signature, sut.GetSignature())
}
