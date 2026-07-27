package resp

// FileUploadResp 文件上传响应
type FileUploadResp struct {
	ID           string `json:"id"`
	FileName     string `json:"file_name"`
	Size         int64  `json:"size"`
	MIMEType     string `json:"mime_type"`
	AccessURL    string `json:"access_url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// FileInfoResp 文件信息响应
type FileInfoResp struct {
	ID             string `json:"id"`
	FileName       string `json:"file_name"`
	Size           int64  `json:"size"`
	MIMEType       string `json:"mime_type"`
	StorageChannel string `json:"storage_channel"`
	MD5Hash        string `json:"md5_hash"`
	AttachType     string `json:"attach_type,omitempty"`
	AttachID       string `json:"attach_id,omitempty"`
	UploaderID     string `json:"uploader_id"`
	ThumbnailURL   string `json:"thumbnail_url,omitempty"`
	AccessURL      string `json:"access_url"`
	CreatedAt      string `json:"created_at"`
}