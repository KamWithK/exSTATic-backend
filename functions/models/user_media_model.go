package models

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
