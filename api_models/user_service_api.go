package api_models

// UserRegisterReqAPI 用户注册请求API
type UserRegisterReqAPI struct {
	// 登录时为form表单
	Password  string `form:"password"`
	Telephone string `form:"telephone"`
	Username  string `form:"username"`
}

// UserLoginReqAPI 用户登录请求API
// todo: 比较疑惑的是为什么要用form才行，而这两个参数是query请求
type UserLoginReqAPI struct {
	Telephone string `form:"telephone"`
	Password  string `form:"password"`
}

// UserLoginRespAPI 用户登录响应API
type UserLoginRespAPI struct {
	Token string `json:"token"`
}

// UserCheckLoginRespAPI 检查登录状态API
type UserCheckLoginRespAPI struct {
	UserId   string `json:"userId"`
	UserName string `json:"username"`
}
