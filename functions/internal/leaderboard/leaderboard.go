package leaderboard

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
