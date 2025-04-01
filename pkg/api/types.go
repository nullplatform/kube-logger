package api

type Config struct {
	Namespace     string
	ApplicationID string
	ScopeID       string
	DeploymentID  string
	Limit         int
	NextPageToken string
	FilterPattern string
	StartTime     string
}

type Result struct {
	Results       []LogEntry `json:"results"`
	NextPageToken string     `json:"nextPageToken"`
}

type LogEntry struct {
	Message string `json:"message"`
}
