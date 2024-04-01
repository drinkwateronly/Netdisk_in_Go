package response

var (
	Success = 0

	Unauthorized     = 10000
	UserNotExist     = 10001 // 用户不存在
	WrongPassword    = 10002
	CookieGenError   = 10003
	CookieNotValid   = 10004
	StorageNotEnough = 10005

	ParentNotExist = 20000 //
	FileRepeat     = 20001
	FileNotExist   = 20002

	FolderLoopError = 30000

	FileRecordNotExist = 20003
	FileNameNotValid   = 20009

	FileIOError      = 30001
	SaveFileNotExist = 30002
	GenZipError      = 30003

	FileSaveError = 50001

	FileNotDeleted       = 60000
	RecoveryFileNotExist = 600001

	ShareExpired           = 60000
	ExtractionCodeNotValid = 60001

	ReqParamNotValid = 10000

	DatabaseError = 99999

	NotSupport = 88888
)
