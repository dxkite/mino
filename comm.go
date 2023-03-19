package mino

// 更新结构体
type UpdateInfo struct {
	Version     string `json:"version"`
	Os          string `json:"os"`
	Arch        string `json:"arch"`
	Binary      string `json:"binary,omitempty"`
	DownloadUrl string `json:"download_url"`
	Message     string `json:"message,omitempty"`
}
