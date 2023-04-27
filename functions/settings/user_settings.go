package user_settings

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

type TableKey struct {
	Username  string `json:"username" binding:"required"`
	MediaType string `json:"media_type"`
}
