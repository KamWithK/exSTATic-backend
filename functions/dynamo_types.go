package dynamo_types

type UserSettingsKey struct {
	Username  string `json:"username" binding:"required"`
	MediaType string `json:"media_type"`
}

type UserSettings struct {
	Username            string   `json:"username" binding:"required"`
	MediaType           string   `json:"media_type"`
	ShowOnLeaderboard   *bool    `json:"show_on_leaderboard"`
	InterfaceBlurAmount *float32 `json:"interface_blur_amount"`
	MenuBlurAmount      *float32 `json:"menu_blur_amount"`
	MaxAFKTime          *float32 `json:"max_afk_time"`
	MaxBlurTime         *float32 `json:"max_blur_time"`
	MaxLoadLines        *int16   `json:"max_load_lines"`
}

type UserMediaKey struct {
	Username        string `json:"username" binding:"required"`
	MediaType       string `json:"media_type" binding:"required"`
	MediaIdentifier string `json:"media_identifier"`
}

type UserMediaEntry struct {
	Username        string  `json:"username" binding:"required"`
	MediaType       string  `json:"media_type" binding:"required"`
	MediaIdentifier string  `json:"media_identifier"`
	DisplayName     *string `json:"display_type"`
	LastUpdate      *string `json:"last_update"`
}

type UserMediaStat struct {
	Username        string  `json:"username" binding:"required"`
	MediaType       string  `json:"media_type" binding:"required"`
	MediaIdentifier string  `json:"media_identifier"`
	Date            *string `json:"date"`
	Stats           *string `json:"stats"`
	LastUpdate      *string `json:"last_update"`
}
