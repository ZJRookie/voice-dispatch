package response

type SoundNotifyResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Detail  string `json:"detail"`
}
