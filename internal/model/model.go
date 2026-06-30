package model

import "time"

type Settings struct {
	QBURL               string `json:"qbUrl"`
	QBUsername          string `json:"qbUsername"`
	QBPassword          string `json:"-"`
	DownloadRoot        string `json:"downloadRoot"`
	DefaultCategory     string `json:"defaultCategory"`
	RSSInterval         int    `json:"rssInterval"`
	DefaultExcludeRegex string `json:"defaultExcludeRegex"`
	LatestExcludeRegex  string `json:"latestExcludeRegex"`
}

type SettingsResponse struct {
	QBURL               string `json:"qbUrl"`
	QBUsername          string `json:"qbUsername"`
	PasswordSet         bool   `json:"passwordSet"`
	DownloadRoot        string `json:"downloadRoot"`
	DefaultCategory     string `json:"defaultCategory"`
	RSSInterval         int    `json:"rssInterval"`
	DefaultExcludeRegex string `json:"defaultExcludeRegex"`
	LatestExcludeRegex  string `json:"latestExcludeRegex"`
}

type UpdateSettingsRequest struct {
	QBURL               string `json:"qbUrl"`
	QBUsername          string `json:"qbUsername"`
	QBPassword          string `json:"qbPassword,omitempty"`
	DownloadRoot        string `json:"downloadRoot"`
	DefaultCategory     string `json:"defaultCategory"`
	RSSInterval         int    `json:"rssInterval"`
	DefaultExcludeRegex string `json:"defaultExcludeRegex"`
}

type Subscription struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	RawTitle     string    `json:"rawTitle"`
	RSSURL       string    `json:"rssUrl"`
	Regex        string    `json:"regex"`
	ExcludeRegex string    `json:"excludeRegex"`
	SaveDirName  string    `json:"saveDirName"`
	SavePath     string    `json:"savePath"`
	RuleName     string    `json:"ruleName"`
	Season       int       `json:"season"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateSubscriptionRequest struct {
	RSSURL        string `json:"rssUrl"`
	Regex         string `json:"regex,omitempty"`
	ExcludeRegex  string `json:"excludeRegex,omitempty"`
	CustomDirName string `json:"customDirName,omitempty"`
	Season        int    `json:"season"`
}

type UpdateSubscriptionRequest struct {
	RSSURL       string `json:"rssUrl"`
	Regex        string `json:"regex,omitempty"`
	ExcludeRegex string `json:"excludeRegex,omitempty"`
	SaveDirName  string `json:"saveDirName"`
	Season       int    `json:"season"`
	Enabled      bool   `json:"enabled"`
}

type QBTestResponse struct {
	Connected     bool   `json:"connected"`
	Version       string `json:"version,omitempty"`
	WebAPIVersion string `json:"webApiVersion,omitempty"`
}

type LogResponse struct {
	Lines []string `json:"lines"`
}
