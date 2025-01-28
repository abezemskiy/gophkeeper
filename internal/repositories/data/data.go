package data

import "time"

const (
	PASSWORD = iota
	TEXT
	BINARY
	BANKCARD
)

const (
	NEW         = iota // статус означает, что добавляются новые данные
	SAVED              // статус, указывающий, что данные успешно сохранены в хранилище
	CHANGE             // статус, указывающий, что данные должны изменить существующую запись
	CONFLICT           // статус, указывающий, что хранящиеся данные находятся в конфликтном состоянии
	FIXCONFLICT        // статус, который означает, что текущее изменение должно разрешить существующий конфликт
)

// Data - структура для передачи данных (пароли, банковские карты) и метаинформации между сервером и клиентом.
type Data struct {
	Data       []byte    `json:"data"`                // поле для хранения полезной нагрузки в виде слайса байт
	Type       int       `json:"type"`                // тип передаваемых данных (пара логин-пароль, банковская карта, бинарные данные и т.д.
	Name       string    `json:"name"`                // уникальное имя сохраняемых данных
	Metainfo   string    `json:"metainfo"`            // произвольная текстовая метаинформация
	Status     int       `json:"status,omitempty"`    // статус данных. Новые данные, сохранены на сервере, изменены
	CreateDate time.Time `json:"create_data"`         // дата создания данных
	EditDate   time.Time `json:"edit_date,omitempty"` // дата редактирования данных
}

// EncryptedData - структура зашифрованных данных для хранения в базе данных.
type EncryptedData struct {
	EncryptedData []byte    `json:"encrypted_data"`      // поле для хранения зашифрованной полезной нагрузки в виде слайса байт
	Name          string    `json:"name"`                // уникальное имя сохраняемых данных
	CreateDate    time.Time `json:"create_data"`         // дата создания данных
	EditDate      time.Time `json:"edit_date,omitempty"` // дата редактирования данных
}
