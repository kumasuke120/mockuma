package myos

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitWd(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	assert.Empty(theWd)
	err := InitWd()
	if assert.Nil(err) {
		assert.NotEmpty(theWd)
	}
}

func TestSetWd(t *testing.T) {
	setWd("TestSetWd")
	assert.Equal(t, "TestSetWd", theWd)
}

func TestGetWd(t *testing.T) {
	setWd("TestGetWd")
	assert.Equal(t, "TestGetWd", theWd)
	assert.Equal(t, theWd, GetWd())
}

//noinspection GoImportUsedAsName
func TestChdir(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	wd, err := os.Getwd()
	require.Nil(err)

	for i := 0; i < 2; i++ {
		err = Chdir(wd)
		if assert.Nil(err) {
			assert.Equal(wd, theWd)
		}
	}
}
