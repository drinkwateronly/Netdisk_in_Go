package api

// UserRegisterReq 用户注册请求API
type UserRegisterReq struct {
	// 登录时为form表单
	Telephone string `form:"telephone"` // 电话
	Password  string `form:"password"`  // 密码
	Username  string `form:"username"`  // 用户名
}

// ----------------------------------------------------------------

// UserLoginReq 用户登录请求API
// todo: 比较疑惑的是为什么要用form才行，而这两个参数是query请求
type UserLoginReq struct {
	Telephone string `form:"telephone"` // 电话（相当于账号）
	Password  string `form:"password"`  // 密码
}

// UserLoginResp 用户登录响应API
type UserLoginResp struct {
	Token string `json:"token"` // cookie
}

// ----------------------------------------------------------------

// UserCheckLoginResp 检查登录状态响应
type UserCheckLoginResp struct {
	UserId   string `json:"userId"`   // 用户id
	UserName string `json:"username"` // 用户名
}
