package models

import (
	"errors"
	"strconv"
	"strings"
)

type UserMediaKey struct {
	Username        string `json:"username" binding:"required"`
	MediaType       string `json:"media_type" binding:"required"`
	MediaIdentifier string `json:"media_identifier"`
}

type UserMediaDateKey struct {
	Key      UserMediaKey `json:"key" binding:"required"`
	DateTime int64        `json:"datetime" binding:"required"`
}

type MediaStat struct {
	TimeRead  int64 `json:"time_read" binding:"required"`
	CharsRead int64 `json:"chars_read" binding:"required"`
	LinesRead int64 `json:"lines_read"`
}

type UserMediaEntry struct {
	Key         UserMediaKey `json:"key" binding:"required"`
	DisplayName string       `json:"display_name"`
	Series      string       `json:"series"`
	LastUpdate  int64        `json:"last_update"`
}

type UserMediaStat struct {
	Key        UserMediaKey `json:"key" binding:"required"`
	Date       *int64       `json:"date"`
	Stats      MediaStat    `json:"stats"`
	LastUpdate int64        `json:"last_update"`
	Pause      bool         `json:"pause"`
}

func UserMediaPK(key UserMediaKey) string {
	return key.MediaType + "#" + key.Username
}

func splitUserMediaPK(pk string, key *UserMediaKey) error {
	if pk == "" {
		return errors.New("empty partition key (pk)")
	}

	pkSplit := strings.Split(pk, "#")

	if len(pkSplit) != 2 || pkSplit[0] == "" || pkSplit[1] == "" {
		return errors.New("invalid partition key (pk) split")
	}

	key.Username = pkSplit[1]
	key.MediaType = pkSplit[0]

	return nil
}

func splitUserMediaSK(sk string, key *UserMediaKey, recordDate *int64) error {
	if sk == "" {
		return errors.New("empty secondary key (sk)")
	}

	skSplit := strings.Split(sk, "#")

	if len(skSplit) == 1 {
		key.MediaIdentifier = skSplit[0]
	} else if len(skSplit) == 2 {
		key.MediaIdentifier = skSplit[1]
		date, parseIntErr := strconv.ParseInt(skSplit[0], 10, 64)

		if parseIntErr != nil {
			return errors.New("could not parse Unix epoch")
		}

		*recordDate = date
	} else {
		return errors.New("invalid secondary key (sk) split")
	}

	return nil
}

func SplitUserMediaCompositeKey(pk string, sk string) (*UserMediaKey, *int64, error) {
	key := UserMediaKey{}
	recordDate := new(int64)

	if pkErr := splitUserMediaPK(pk, &key); pkErr != nil {
		return nil, nil, pkErr
	}
	if skErr := splitUserMediaSK(sk, &key, recordDate); skErr != nil {
		return nil, nil, skErr
	}

	return &key, recordDate, nil
}
