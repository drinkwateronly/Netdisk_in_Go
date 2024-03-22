package response

var (
	Success = 0

	Unauthorized   = 10000
	UserNotExist   = 10001 // 用户不存在
	WrongPassword  = 10002
	CookieGenError = 10003
	CookieNotValid = 10004

	PARENTNOTEXIST = 20000 //
	FILEREPEAT     = 20001
	FileNotExist   = 20002

	FileRecordNotExist = 20003
	FileNameNotValid   = 20009

	FILECREATEERROR    = 50000
	FILESAVEERROR      = 50001
	FILETYPENOTSUPPORT = 50002

	ShareExpired           = 60000
	ExtractionCodeNotValid = 60001

	ReqParamNotValid = 10000

	DATABASEERROR = 99999
	DatabaseError = 99999

	NotSupport = 88888
)
