package utils

type CompositeKey struct {
	PK interface{} `json:"pk" binding:"required"`
	SK interface{} `json:"sk"`
}

type UserSettingsKey struct {
	Username  string `json:"username" binding:"required"`
	MediaType string `json:"media_type"`
}

type UserSettings struct {
	Key                 UserSettingsKey `json:"key" binding:"required"`
	ShowOnLeaderboard   *bool           `json:"show_on_leaderboard"`
	InterfaceBlurAmount *float32        `json:"interface_blur_amount"`
	MenuBlurAmount      *float32        `json:"menu_blur_amount"`
	MaxAFKTime          *int16          `json:"max_afk_time"`
	MaxBlurTime         *int16          `json:"max_blur_time"`
	MaxLoadLines        *int16          `json:"max_load_lines"`
}

type UserMediaKey struct {
	Username        string `json:"username" binding:"required"`
	MediaType       string `json:"media_type" binding:"required"`
	MediaIdentifier string `json:"media_identifier"`
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

type LeaderboardKey struct {
	Username   string `json:"username" binding:"required"`
	TimePeriod string `json:"time_period" binding:"required"`
	MediaType  string `json:"media_type" binding:"required"`
}

type LeaderboardEntry struct {
	Key        LeaderboardKey `json:"key" binding:"required"`
	MediaNames string         `json:"media_names"`
	TimeRead   int64          `json:"time_read"`
	CharsRead  int64          `json:"chars_read"`
}
