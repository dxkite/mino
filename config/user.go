package config

type UserConfig struct {
	Enable   bool       `yaml:"enable"`
	UserList []UserItem `yaml:"user_list"`
}

type UserItem struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Freeze   bool   `yaml:"freeze"`
}
