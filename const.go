package mino

// 版本号
const Version = "0.1.4-alpha"

// 默认自动更新网址
const Upload = ""

// 可运行机器
const MachineId = ""

const (
	KeyUsername        = "username"
	KeyPassword        = "password"
	KeyAddress         = "address"
	KeyPacFile         = "pac_file"
	KeyUpstream        = "upstream"
	KeyInput           = "input"
	KeyDataPath        = "data_path"
	KeyMaxStreamRewind = "max_stream_rewind"
	KeyWebRoot         = "web_root"
	KeyAutoStart       = "auto_start"
	KeyAutoUpdate      = "auto_update"
	KeyLogFile         = "log_file"
	KeyLogLevel        = "log_level"
	KeyConfFile        = "conf_file"
	KeyRootCa          = "tls_root_ca"
	KeyCertFile        = "tls_cert_file"
	KeyKeyFile         = "tls_key_file"
	KeyEncoder         = "encoder"
	KeyUpdateUrl       = "update_url"
	KeyLatestVersion   = "latest_version"
	KeyXorMod          = "xor_mod"
	KeyDump            = "dump"

	KeyWebAuth        = "web_auth"
	KeyWebUsername    = "web_username"
	KeyWebPassword    = "web_password"
	KeyWebFailedTimes = "web_failed_times"
)

const PathMinoPac = "/mino.pac"
