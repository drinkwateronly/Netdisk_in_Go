package officemodels

type Info struct {
	Owner  string `json:"owner"`  // "Me"即可
	Upload string `json:"upload"` // "上传的时间字符串"
	//Favorite string `json:"favorite"`
}

type Permission struct {
	Comment              bool              `json:"comment" default:"true"`
	Copy                 bool              `json:"copy" default:"true"`
	Download             bool              `json:"download" default:"true"`
	Edit                 bool              `json:"edit"`
	Print                bool              `json:"print" default:"true"`
	FillForms            bool              `json:"fillForms" default:"true"`
	ModifyFilter         bool              `json:"modifyFilter" default:"true"`
	ModifyContentControl bool              `json:"modifyContentControl" default:"true"`
	Review               bool              `json:"review" default:"true"`
	Chat                 bool              `json:"chat" default:"true"`
	CommentGroup         map[string]string `json:"commentGroup"`
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

// Logo the image file at the top left corner of the Editor header
type Logo struct {
	Image         string      `json:"image"`         // the path to the image file used to show in common work mode
	ImageDark     string      `json:"imageDark"`     //
	ImageEmbedded interface{} `json:"imageEmbedded"` // the path to the image file used to show in the emb
	Url           string      `json:"url"`           // the absolute URL which will be used when someone clicks the
}

type Customization struct {
	Logo               Logo        `json:"logo"`
	AutoSave           bool        `json:"autosave" default:"true"`
	Comments           bool        `json:"comments" default:"true"`
	CompactHeader      bool        `json:"compactHeader" default:"true"`
	CompactToolbar     interface{} `json:"compactToolbar" default:"true"`
	CompatibleFeatures interface{} `json:"compatibleFeatures" default:"true"`
	ForceSave          interface{} `json:"forcesave" default:"true"`
	Help               bool        `json:"help" default:"true"`
	HideNotes          interface{} `json:"hideNotes"`
	HideRightMenu      interface{} `json:"hideRightMenu" default:"false"`
	HideRulers         interface{} `json:"hideRuler" default:"false"`
	SubmitForm         bool        `json:"submitForm" default:"false"`
	About              bool        `json:"about" default:"true"`
	FeedBack           bool        `json:"feedBack" default:"true"`
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
	Lang          string        `json:"lang" default:"zh"`
	Mode          string        `json:"mode" default:"view"`
	User          User          `json:"user"`
	Templates     []Template    `json:"templates"`
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
	DocserviceApiUrl string `json:"docserviceApiUrl" default:"https://172.31.226.34:9696/web-apps/apps/api/documents/api.js"`
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
		Chat:                 true,
		CommentGroup:         nil,
	}
	DefaultCustomization = Customization{
		Logo:               Logo{},
		AutoSave:           true,
		Comments:           true,
		CompactHeader:      true,
		CompactToolbar:     true,
		CompatibleFeatures: true,
		ForceSave:          true,
		Help:               true,
		HideRightMenu:      false,
		HideRulers:         false,
		SubmitForm:         false,
		About:              true,
		FeedBack:           true,
	}
	DefaultCoEditing = CoEditing{
		Mode:   "fast",
		Change: true,
	}
	PreviewUrlFormat = "http://172.31.226.81:8080/office/preview?userFileId=%s&token=%s"
)

func NewData(user User, token string, document Document, documentType string) *Data {
	data := Data{
		File: File{
			Document:     document,
			DocumentType: documentType,
			EditorConfig: EditorConfig{
				Customization: DefaultCustomization,
				CallbackUrl:   "http://172.31.226.81:8080/office/callback",
				CreateUrl:     "http://172.31.226.81:8080/file/createFile",
				Lang:          "zh",
				Mode:          "edit",
				CoEditing:     DefaultCoEditing,
				User:          user,
			},
			Token: token,
			Type:  "desktop",
		},
		DocserviceApiUrl: "http://172.31.226.34:9696/web-apps/apps/api/documents/api.js",
		ReportName:       "123",
	}

	return &data
}

type OfficeError struct {
	Error int `json:"error"`
}
