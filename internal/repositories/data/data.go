package data

const (
	PASSWORD = iota
	TEXT
	BINARY
	BANKCARD
)

const (
	NEW = iota
	CHANGED
	SAVED
)

// Data - структура для передачи данных (пароли, банковские карты) и метаинформации между сервером и клиентом.
type Data struct {
	Data     []byte `json:"data"`     // поле для хранения полезной нагрузки в виде слайса байт
	Type     int    `json:"type"`     // тип передаваемых данных (пара логин-пароль, банковская карта, бинарные данные и т.д.
	Name     string `json:"name"`     // уникальное имя сохраняемых данных
	Metainfo string `json:"metainfo"` // произвольная текстовая метаинформация
	Status   int    `json:"status"`   // статус данных. Новые данные, сохранены на сервере, изменены
}
