package api_models

var (
	Success = 0

	USERNOTEXIST = 10001 // 用户不存在
	UNAUTHORIZED = 10000

	PARENTNOTEXIST = 20000 //
	FILEREPEAT     = 20001
	FILENOTEXIST   = 20002

	FileRecordNotExist = 20003

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
