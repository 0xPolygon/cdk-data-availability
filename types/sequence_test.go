package types

import (
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSequenceCase struct {
	s            Sequence
	expectedHash common.Hash
}

const data = "0b000015bc000000000b00000002000000000b00000002000000000b0000000200000000e708843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e98080eb60de9d2ae8d392e2a7f974b7f35bb7cdfc4f32f8f69212bc426d027fb1e6e13fbdcdc6cc9a7427fd60e2a108464d753d4b0c422fea68b467be60d693c37a9a1cffe709843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e98080e5ef8cee91269b4e5b482217026cb966d7d89e6aa0bfa1b8184601261973147c6c0b1e270f90c2f598a5bc59c99f6da6e3bf1d07cd02a7e3a973b16c40b20d641bffe70a843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e980802cf438a125eec0bedc70af154feaa04a0fb940844fcee1735f2c0a29a1bbb2444c28556ac3d9ab99f006a1af45d3c69c7134c10836cd8e10eb0c4a0536be12481cffe70b843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e980809abd370489d86473c8581d30c22096067029a783c0079292a43a039c186c8df86f59050aa02cdec4b04f1c59502897c9628977f7cfec35e2a93875d21e54bdc61bffe70c843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e9808073daf8314747e83d3bd389287978cd723d99077777c209647a9cf06ce2d6fb433065027abcd9ca05cb1d808c11bdb2b2d152fde6ec06dffe1e7b33627100c0011cffe70d843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e98080c46a74f28c0d25b647276ea3c14779a67516c37be1aeed655b53cb955b32badb3952986001056f0b1b902c3cab73a62e20195f917ce4c26f6faa7e5e60deacb21cffe70e843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e98080dd3ca0dd2aca53110c8324a0ce09d69db4b791c2bbfff7520c892d2c943c9fec126678b24364fb296ce3326d7a7de446c14bbab6a563e008d96e0859614500dd1bffe70f843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e98080216f364ba880322afa0a9f8c507c1e0dad7cc7a7cbfcccb05f8f4b6acdd0f21a5da25704bba861e29691945fb76875ec9dff27cefce58f58cac6569e49308e8a1cffe710843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e98080ccd21e98490bb85f00e83d4b00d2d4eb2a31a5f2f71095c76a18090c159ed7c575ed18a90896bc83380af3c332f1896e2236534fed1fb9807e0dcd853f21f1b91cffe711843b9aca008252089470997970c51812dc3a010c7d01b50e0d17dc79c8822710808203e9808002a08668ec08dafa4d6b0a66cb42b02bed3b016c1c156356f478e9163c605b48761f5f7ef1191820ff8246da7afbb2faea8a27211344984d43773ad3e75fb7591cff"

var testSequenceCases = []testSequenceCase{
	{
		s:            Sequence{ArgBytes(common.Hex2Bytes(data))},
		expectedHash: common.BytesToHash(common.Hex2Bytes("f9dab0ab5bce4b4b002bc5f117d8a5357ab6402dfcc10b0514fa5851b6a37ba1")),
	},
}

func TestHashToSign(t *testing.T) {
	for _, c := range testSequenceCases {
		assert.Equal(
			t, c.expectedHash.Hex(),
			"0x"+common.Bytes2Hex(c.s.HashToSign()),
		)
		require.Equal(t, 1, len(c.s.OffChainData()))
	}
}

func TestSigning(t *testing.T) {
	privKeys := []*ecdsa.PrivateKey{}
	for i := 0; i < 5; i++ {
		pk, err := crypto.GenerateKey()
		require.NoError(t, err)
		privKeys = append(privKeys, pk)
	}
	for _, c := range testSequenceCases {
		for _, pk := range privKeys {
			signedSequence, err := c.s.Sign(pk)
			require.NoError(t, err)
			actualAddr, err := signedSequence.Signer()
			require.NoError(t, err)
			expectedAddr := crypto.PubkeyToAddress(pk.PublicKey)
			assert.Equal(t, expectedAddr, actualAddr)
		}
	}
}
