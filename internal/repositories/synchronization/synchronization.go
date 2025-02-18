package synchronization

import "time"

// PeroifOfSynchr - период синхронизации данных между сервером и клиентом.
// Период за который клиент должен отправить серверу запрос на синхронизацию данных.
const PeroidOfSynchr = time.Minute * 1

// GetPeroidOfSynchr - функция для получения периода синхронизации данных между сервером и клиентом.
func GetPeroidOfSynchr() time.Duration {
	return PeroidOfSynchr
}
