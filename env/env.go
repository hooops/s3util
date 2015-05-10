package env

var ACCESS_KEY, SECRET_KEY string

func Init(access, secret string) {
	ACCESS_KEY = access
	SECRET_KEY = secret
}
