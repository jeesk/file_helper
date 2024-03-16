package db

type Message struct {
	Id          int64  `json:"id" gorm:"primary_key;auto_increment"`
	MessageType string `json:"messageType"`
	HeadIcon    string `json:"headIcon"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	HtitleType  string `json:"htitleType"`
	Htitle      string `json:"htitle"`
	Html        string `json:"html"`
}
