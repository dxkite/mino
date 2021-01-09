package mino

// 版本号
const Version = "0.1.4-alpha"

// 默认自动更新网址
const Upload = ""

// 可运行机器
const MachineId = ""

const (
	// upstream 账号密码
	KeyUsername = "username"
	KeyPassword = "password"
	// 监听地址
	KeyAddress = "address"
	// pac文件
	KeyPacFile = "pac_file"
	// 上传流
	KeyUpstream = "upstream"
	// 输入流
	KeyInput = "input"
	// 数据存储位置
	KeyDataPath = "data_path"
	// web服务器根目录
	KeyWebRoot = "web_root"
	// 自动重启(windows)
	KeyAutoStart = "auto_start"
	// 自动更新
	KeyAutoUpdate = "auto_update"
	// 日志文件
	KeyLogFile = "log_file"
	// 日志等级
	KeyLogLevel = "log_level"
	// 配置文件路径
	KeyConfFile = "conf_file"
	// 更新检擦地址
	KeyUpdateUrl = "update_url"
	// 作为更新服务器使用，指明最后版本
	KeyLatestVersion = "latest_version"
	// 加密传输类型，xor/tls 默认不开启
	KeyEncoder = "encoder"
	// xor 长度，默认4
	KeyXorMod = "xor_mod"
	// TLS连接CA
	KeyRootCa = "tls_root_ca"
	// TLS密钥
	KeyCertFile = "tls_cert_file"
	KeyKeyFile  = "tls_key_file"
	// dump 数据流，默认false
	KeyDump = "dump"
	// HTTP预读
	KeyMaxRewindSize = "http_max_rewind_size"
	// 流预读，默认 8
	KeyMaxStreamRewind = "max_stream_rewind"
	// 管理Web服务设置，默认false
	KeyWebAuth     = "web_auth"
	KeyWebUsername = "web_username"
	KeyWebPassword = "web_password"
	// 登录失败次数
	KeyWebFailedTimes = "web_failed_times"
)

const PathMinoPac = "/mino.pac"
