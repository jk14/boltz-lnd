package boltz

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
	"testing"
)

var chainParams = &chaincfg.MainNetParams
var redeemScript, _ = hex.DecodeString("a9146a24b142de20b50871a247c1c66a6e41ee199017876321038ce1d1be5a22b396ccafc109c86717bc081301fe58d1958546d5aba647047af3670381a81ab1752102d23a7d39395f40a71a490cf79e0f2df5da2fb006fdab660bc0c78ef0c9ba457668ac")

func TestCheckSwapAddress(t *testing.T) {
	// P2SH nested P2WSH address
	err := CheckSwapAddress(chainParams, "3F8UixJcrfxCaGpRryyRuKotBFXRFeW7ej", redeemScript, true)
	assert.Nil(t, err)

	err = CheckSwapAddress(chainParams, "", redeemScript, true)
	assert.Equal(t, errors.New("invalid address"), err)

	// P2WSH address
	err = CheckSwapAddress(chainParams, "bc1q73lzkly9le40qxym5wh5wyp0davanw3u9m0u28wafay4ay7z34cscztt48", redeemScript, false)
	assert.Nil(t, err)

	err = CheckSwapAddress(chainParams, "", redeemScript, false)
	assert.Equal(t, errors.New("invalid address"), err)
}

func TestWitnessScriptHashAddress(t *testing.T) {
	address, err := WitnessScriptHashAddress(chainParams, redeemScript)

	assert.Nil(t, err)
	assert.Equal(t, "bc1q73lzkly9le40qxym5wh5wyp0davanw3u9m0u28wafay4ay7z34cscztt48", address)
}

func TestScriptHashAddress(t *testing.T) {
	address, err := ScriptHashAddress(chainParams, redeemScript)

	assert.Nil(t, err)
	assert.Equal(t, "32Hjgh4J1kZFGbuJ9aPwqmqz3L5GkhNAzR", address)
}

func TestNestedScriptHashAddress(t *testing.T) {
	address, err := NestedScriptHashAddress(chainParams, redeemScript)

	assert.Nil(t, err)
	assert.Equal(t, "3F8UixJcrfxCaGpRryyRuKotBFXRFeW7ej", address)
}
