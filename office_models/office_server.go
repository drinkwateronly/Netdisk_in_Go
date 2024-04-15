package office_models

import (
	"fmt"
	"netdisk_in_go/sysconfig"
)

type Info struct {
	Owner  string `json:"owner"`  // "Me"即可
	Upload string `json:"upload"` // "上传的时间字符串"
}

type Permission struct {
	Copy     bool `json:"copy"`
	Download bool `json:"download"`
	Edit     bool `json:"edit"`
	Print    bool `json:"print"`
}

type Document struct {
	FileType    string     `json:"fileType"`
	Info        Info       `json:"info"`
	Key         string     `json:"key"` // key
	Permissions Permission `json:"permissions"`
	Title       string     `json:"title"` // "123.xlsx"
	Url         string     `json:"url"`   //
	UserFileId  string     `json:"userFileId"`
}

type Customization struct {
	ForceSave bool `json:"forcesave"`
}

type User struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

type CoEditing struct {
	Mode   string `json:"mode" default:"fast"`
	Change bool   `json:"change" default:"true"`
}

type EditorConfig struct {
	CallbackUrl   string        `json:"callbackUrl"`
	Customization Customization `json:"customization"`

	Lang   string `json:"lang"`
	Region string `json:"region"`
	Mode   string `json:"mode"`
	User   User   `json:"user"`
}

type File struct {
	Document     Document     `json:"document"`
	DocumentType string       `json:"documentType"`
	EditorConfig EditorConfig `json:"editorConfig"`
}

type OnlyOfficeConfig struct {
	File             File   `json:"file"`
	DocserviceApiUrl string `json:"docserviceApiUrl"`
	ReportName       string `json:"reportName"`
}

var (
	DefaultPermissions = Permission{
		Copy:     true,
		Download: true,
		Edit:     true,
		Print:    true,
	}
	DefaultCustomization = Customization{
		ForceSave: true, // 允许ctrl+s强制保存
	}
	PreviewUrlFormat = "http://172.31.226.81:8080/office/preview?userFileId=%s&token=%s"
)

func NewOnlyOfficeConfig(user User, token string, document Document, documentType string) *OnlyOfficeConfig {
	officeConfig := sysconfig.Config.OfficeConfig
	//backendConfig := sysconfig.Config.
	callbackUrl := "http://172.31.226.81:8080/office/callback"
	data := OnlyOfficeConfig{
		File: File{
			Document:     document,
			DocumentType: documentType,
			EditorConfig: EditorConfig{
				Customization: DefaultCustomization,
				CallbackUrl:   callbackUrl,
				//CreateUrl:   createUrl,
				//Lang:          "zh-CN",
				//Region:        "zh",
				Mode: "edit",
				User: user,
			},

			//Token: token,
			//Type:  "desktop",
		},
		DocserviceApiUrl: fmt.Sprintf("http://%s:%s/web-apps/apps/api/documents/api.js", officeConfig.Host, officeConfig.Port), // office前端显示api
		//ReportName:       "123",
	}

	return &data
}
