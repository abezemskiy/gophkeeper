package binary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/encr"
	"gophkeeper/internal/client/handlers"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/client/storage/data"
	"gophkeeper/internal/client/storage/data/tools/binary"
	"gophkeeper/internal/client/tui/app"
	"gophkeeper/internal/client/tui/tools/printer"
	repoData "gophkeeper/internal/repositories/data"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// dataInfo - вспомогательная структура для передачи полученных от пользоавтеля данных в функцию сохранения данных в сервисе.
type dataInfo struct {
	binary     data.Binary
	path       string
	metaInfo   string
	name       string
	createDate time.Time
	editDate   time.Time
}

// AddBinaryPage - TUI страница добавления нового файла.
func AddBinaryPage(ctx context.Context, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	info identity.IUserInfoStorage, app app.App) tview.Primitive {

	form := tview.NewForm()
	// структура для введенной пары логин пароль
	dataInfo := dataInfo{
		createDate: time.Now(),
		editDate:   time.Now(),
	}

	form.AddInputField("Имя данных", "", 20, nil, func(text string) {
		if text != "" {
			dataInfo.name = text // Разрешаю ввод, только если установлено имя данных
		}
	})
	form.AddInputField("Путь к файлу", "", 20, nil, func(text string) {
		if text != "" {
			dataInfo.path = text // Разрешаю ввод, только если установлен путь к файлу
		}
	})
	form.AddInputField("Описание", "", 20, nil, func(text string) { dataInfo.metaInfo = text })

	form.AddButton("Сохранить", func() {
		// проверяю наличие в приложении мастер пароля
		authData, id := info.Get()
		if authData.Password == "" {
			// мастер пароль не установлен, возвращаю пользователя на страницу аутентификации.
			app.SwitchTo("login")
		}

		// Преобразую файл в бинарный вид
		err := parseFile(&dataInfo)
		if err != nil {
			logger.ClientLog.Error("failed to parse file to binary", zap.String("error", error.Error(err)))
		}

		ok, err := save(ctx, id, url, client, stor, &dataInfo, authData.Password)
		if err != nil {
			logger.ClientLog.Error("save data error", zap.String("error", error.Error(err)))
			printer.Error(app, fmt.Sprintf("save data error, %v", err))
		}
		if !ok {
			logger.ClientLog.Error("data is not unique", zap.String("name", dataInfo.name))
			printer.Error(app, fmt.Sprintf("data is not unique, name %s", dataInfo.name))
		}

		// Печатаю сообщение об успешном сохранении данных
		printer.Message(app, "data saved successfully")

		// перенаправляю пользователя на страницу данных
		app.SwitchTo("home")
	})
	form.AddButton("Отмена", func() { app.SwitchTo("add") })

	form.SetBorder(true).SetTitle("Добавить файл")
	return form
}

// parseFile - функция для преобразования содержимого файла в бинарный вид
func parseFile(dataInfo *dataInfo) error {
	// Определяю тип файла
	fileType, err := binary.GetFileType(dataInfo.path)
	if err != nil {
		return fmt.Errorf("error getting file type, %w", err)
	}
	// Преобразую указанный файл в бинарный вид
	d, err := os.ReadFile(dataInfo.path)
	if err != nil {
		return fmt.Errorf("failed to read file, %w", err)
	}
	// Устанавливаю полученные параметры в переменную
	dataInfo.binary.Type = fileType
	dataInfo.binary.Binary = d
	
	return nil
}

func save(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	dataInfo *dataInfo, masterPass string) (bool, error) {

	// сериализую данные типа "BINARY"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(dataInfo.binary); err != nil {
		return false, fmt.Errorf("encode data error, %w", err)
	}

	// Создаю структуру типа data.Data
	dataToEncr := &repoData.Data{
		Data:       buf.Bytes(),
		Type:       repoData.BINARY,
		Name:       dataInfo.name,
		Metainfo:   dataInfo.metaInfo,
		Status:     repoData.NEW,
		CreateDate: dataInfo.createDate,
		EditDate:   dataInfo.editDate,
	}

	// шифрую данные с помощью мастер пароля пользователя
	encrData, err := encr.EncryptData(masterPass, dataToEncr)
	if err != nil {
		return false, fmt.Errorf("failed to encrypt data, %w", err)
	}

	// Сохраняю данные в хранилище
	ok, err := handlers.SaveEncryptedData(ctx, userID, url, client, stor, encrData)
	if err != nil {
		return false, fmt.Errorf("failed to save data, %w", err)
	}
	if !ok {
		return false, nil
	}
	return true, nil
}
