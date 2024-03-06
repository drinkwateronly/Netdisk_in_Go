package office_models

type CallbackHandler struct {
	Actions       []map[string]string `json:"actions"`        // actions:[map[type:1 userid:001]]
	ChangeHistory []interface{}       `json:"changeshistory"` //
	ChangesURL    string              `json:"changesurl"`
	FileType      string              `json:"filetype"`
	ForceSaveType int                 `json:"forcesavetype"`
	History       interface{}         `json:"history"`
	Key           string              `json:"key"`
	Status        int                 `json:"status"`
	Url           string              `json:"url"`
	UserData      string              `json:"userdata"`
	Users         []string            `json:"users"`
}
