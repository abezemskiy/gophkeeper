package data

// Password - структура для хранения пары логин/пароль.
type Password struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Text - структура для хранения произвольных текстовых данных.
type Text struct {
	Text string `json:"text"`
}

// Binary - структура для хранения произвольных
type Binary struct {
	Binary []byte `json:"binary"`
}

// Bank - структура для хранения бинарных данных.
type Bank struct {
	Number int64  `json:"number"`
	Mounth int    `json:"month"`
	Year   int    `json:"year"`
	CVV    int    `json:"cvv"`
	Owner  string `json:"owner"`
}
