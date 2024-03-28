package office_models

import (
	"fmt"
	"netdisk_in_go/sysconfig"
)

type Info struct {
	Owner  string `json:"owner"`  // "Me"即可
	Upload string `json:"upload"` // "上传的时间字符串"
	//Favorite string `json:"favorite"`
}

type Permission struct {
	Chat                 bool              `json:"chat" default:"true"`
	Comment              bool              `json:"comment" default:"true"`
	CommentGroup         map[string]string `json:"commentGroups"`
	Copy                 bool              `json:"copy" default:"true"`
	Download             bool              `json:"download" default:"true"`
	Edit                 bool              `json:"edit"`
	Print                bool              `json:"print" default:"true"`
	FillForms            bool              `json:"fillForms" default:"true"`
	ModifyFilter         bool              `json:"modifyFilter" default:"true"`
	ModifyContentControl bool              `json:"modifyContentControl" default:"true"`
	Review               bool              `json:"review" default:"true"`
}

type Document struct {
	FileType    string     `json:"fileType"`
	Info        Info       `json:"info"`
	Key         string     `json:"key"` // key
	Permissions Permission `json:"permissions"`
	Title       string     `json:"title"` // "123.xlsx"
	Url         string     `json:"url"`   // "https://netdisk.qiwenshare.com:443/filetransfer/preview?userFileId=1742499406193123328&isMin=false&shareBatchNum=undefined&extractionCode=undefined&
	UserFileId  string     `json:"userFileId"`
}

type Goback struct {
}

// Logo the image file at the top and left corner of the Editor header
type Logo struct {
	Image         string      `json:"image"`         // the path to the image file used to show in common work mode
	ImageDark     string      `json:"imageDark"`     //
	ImageEmbedded interface{} `json:"imageEmbedded"` // the path to the image file used to show in the emb
	Url           string      `json:"url"`           // the absolute URL which will be used when someone clicks the
}

type Customization struct {
	//Logo               Logo        `json:"logo"`
	AutoSave           bool `json:"autosave" default:"true"`
	Comments           bool `json:"comments" default:"true"`
	CompactHeader      bool `json:"compactHeader" default:"true"`
	CompactToolbar     bool `json:"compactToolbar" default:"true"`
	CompatibleFeatures bool `json:"compatibleFeatures" default:"true"`
	//ForceSave          interface{} `json:"forcesave" default:"true"`
	//Help          bool        `json:"help" default:"true"`

}

type Embedded struct {
	EmbedUrl      string `json:"embedUrl"`
	SaveUrl       string `json:"saveUrl"`
	ShareUrl      string `json:"shareUrl"`
	ToolbarDocked string `json:"toolbarDocked"`
}

type Template struct {
	Image string `json:"image"`
	Title string `json:"title"`
	Url   string `json:"url"`
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
	ActionLink    interface{}   `json:"actionLink"` // nil
	CallbackUrl   string        `json:"callbackUrl"`
	CoEditing     CoEditing     `json:"coEditing"`
	CreateUrl     string        `json:"createUrl"`
	Customization Customization `json:"customization"`

	HideNotes     bool `json:"hideNotes"`
	HideRightMenu bool `json:"hideRightMenu" default:"false"`
	HideRulers    bool `json:"hideRuler" default:"false"`
	SubmitForm    bool `json:"submitForm" default:"false"`

	Lang      string     `json:"lang" default:"zh"`
	Region    string     `json:"region" default:""`
	Mode      string     `json:"mode" default:"view"`
	User      User       `json:"user"`
	Templates []Template `json:"templates"`
	ForceSave bool       `json:"forcesave"`
}

type File struct {
	Document     Document     `json:"document"`
	DocumentType string       `json:"documentType"`
	EditorConfig EditorConfig `json:"editorConfig"`
	Token        string       `json:"token"`
	Type         string       `json:"type"`
}

type Data struct {
	File             File   `json:"file"`
	DocserviceApiUrl string `json:"docserviceApiUrl"`
	ReportName       string `json:"reportName"`
}

var (
	DefaultPermissions = Permission{
		Comment:              true,
		Copy:                 true,
		Download:             true,
		Edit:                 true,
		Print:                true,
		FillForms:            true,
		ModifyFilter:         true,
		ModifyContentControl: true,
		Review:               true,
		Chat:                 false,
		CommentGroup:         make(map[string]string, 0),
	}
	DefaultCustomization = Customization{
		AutoSave:           true,
		Comments:           true,
		CompactHeader:      true,
		CompactToolbar:     true,
		CompatibleFeatures: true,
	}
	DefaultCoEditing = CoEditing{
		Mode:   "fast",
		Change: true,
	}
	PreviewUrlFormat = "http://172.31.226.81:8080/office/preview?userFileId=%s&token=%s"
)

func NewData(user User, token string, document Document, documentType string) *Data {
	config := sysconfig.Config.OfficeConfig

	callbackUrl := "http://172.31.226.81:8080/office/callback"
	createUrl := "http://172.31.226.81:8080/file/createFile"
	docserviceApiUrl := fmt.Sprintf("http://%s:%s/web-apps/apps/api/documents/api.js", config.Host, config.Port)
	data := Data{
		File: File{
			Document:     document,
			DocumentType: documentType,
			EditorConfig: EditorConfig{
				Customization: DefaultCustomization,
				CallbackUrl:   callbackUrl,
				CreateUrl:     createUrl,

				HideNotes:     false,
				HideRulers:    false,
				HideRightMenu: false,
				SubmitForm:    true,

				Lang: "zh-CN",
				//Region:      "zh",
				Mode:      "edit",
				ForceSave: false,
				CoEditing: DefaultCoEditing,
				User:      user,
			},

			Token: token,
			Type:  "desktop",
		},
		DocserviceApiUrl: docserviceApiUrl,
		//ReportName:       "123",
	}

	return &data
}
