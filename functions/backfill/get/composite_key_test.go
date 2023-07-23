package main

import (
	"testing"

	"github.com/KamWithK/exSTATic-backend/models"
	"github.com/KamWithK/exSTATic-backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestNoKeys(t *testing.T) {
	key, date, err := models.SplitUserMediaCompositeKey("", "")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)
}

func TestNoPK(t *testing.T) {
	key, date, err := models.SplitUserMediaCompositeKey("", "identifier")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)
}

func TestNoSK(t *testing.T) {
	key, date, err := models.SplitUserMediaCompositeKey("type#username", "")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)
}

func TestEmptyPKFields(t *testing.T) {
	key, date, err := models.SplitUserMediaCompositeKey("#", "identifier")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)

	key, date, err = models.SplitUserMediaCompositeKey("a#", "identifier")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)

	key, date, err = models.SplitUserMediaCompositeKey("#a", "identifier")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)
}

func TestEmptySKFields(t *testing.T) {
	key, date, err := models.SplitUserMediaCompositeKey("type#username", "#")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)

	key, date, err = models.SplitUserMediaCompositeKey("type#username", "a#")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)

	key, date, err = models.SplitUserMediaCompositeKey("type#username", "#a")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)
}

func TestInvalidDate(t *testing.T) {
	key, date, err := models.SplitUserMediaCompositeKey("type#username", "a#identifier")
	assert.Empty(t, key)
	assert.Nil(t, date)
	assert.Error(t, err)
}

func TestValidDates(t *testing.T) {
	key, date, err := models.SplitUserMediaCompositeKey("type#username", "0#identifier")
	assert.NotEmpty(t, key)
	assert.EqualValues(t, *date, 0)
	assert.NoError(t, err)

	key, date, err = models.SplitUserMediaCompositeKey("type#username", "1#identifier")
	assert.NotEmpty(t, key)
	assert.EqualValues(t, *date, 1)
	assert.NoError(t, err)

	key, date, err = models.SplitUserMediaCompositeKey("type#username", utils.ZeroPadInt64(0)+"#identifier")
	assert.NotEmpty(t, key)
	assert.EqualValues(t, *date, 0)
	assert.NoError(t, err)

	key, date, err = models.SplitUserMediaCompositeKey("type#username", utils.ZeroPadInt64(101)+"#identifier")
	assert.NotEmpty(t, key)
	assert.EqualValues(t, *date, 101)
	assert.NoError(t, err)
}
