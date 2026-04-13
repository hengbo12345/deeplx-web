package models

type TranslateRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type TranslateResponse struct {
	Success bool   `json:"success"`
	Data    struct {
		Result      string   `json:"result"`
		ID          int      `json:"id,omitempty"`
		Alternatives []string `json:"alternatives,omitempty"`
	} `json:"data,omitempty"`
	Error *ErrorInfo `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Language struct {
	Code string
	Name string
}

var Languages = []Language{
	{"auto", "自动检测"},
	{"ZH", "中文"},
	{"EN", "英语"},
	{"JA", "日语"},
	{"KO", "韩语"},
	{"FR", "法语"},
	{"DE", "德语"},
	{"ES", "西班牙语"},
	{"IT", "意大利语"},
	{"PT", "葡萄牙语"},
	{"RU", "俄语"},
}