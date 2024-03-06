package ApiModels

type UserRegisterApi struct {
	Password  string `form:"password"`
	Telephone string `form:"telephone"`
	Username  string `form:"username"`
}
