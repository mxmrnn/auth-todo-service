package dto

type UserTaskStats struct {
	AsAssignee int `json:"as_assignee"`
	AsWatcher  int `json:"as_watcher"`
}
