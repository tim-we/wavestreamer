package webapp

import "github.com/tim-we/wavestreamer/player"

type ApiNowResponse struct {
	Status      string              `json:"status"`
	Now         *ApiNowPlayingEvent `json:"now"`
	LibraryInfo ApiNowLibraryInfo   `json:"library"`
	Uptime      string              `json:"uptime"`
}

type ApiNowPlayingEvent struct {
	Current string                `json:"current"`
	IsPause bool                  `json:"isPause"`
	History []player.HistoryEntry `json:"history"`
}

type ApiNowLibraryInfo struct {
	Music int `json:"music"`
	Hosts int `json:"hosts"`
	Other int `json:"other"`
	Night int `json:"night"`
}

type ApiOkResponse struct {
	Status string `json:"status"`
}

type ApiErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ApiSearchResponse struct {
	Status  string              `json:"status"`
	Results []SearchResultEntry `json:"results"`
}

type SearchResultEntry struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ApiConfigResponse struct {
	Status string `json:"status"`
	News   bool   `json:"news"`
}
