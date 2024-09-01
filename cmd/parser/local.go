package main

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

// createLocalDir создает локальную директорию
func createLocalDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// prepareLocalDir подгатавливает локальные директории
func prepareLocalDirs() error {
	var paths = [3]string{"zip", "error", "done"}

	slog.Info("Выбран локальный тип источника")

	rootPath := config.GetString("source.local.rootPath")
	for _, path := range paths {
		if err := createLocalDir(filepath.Join(rootPath, path)); err != nil {
			return err
		}
	}

	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Корневая директория источника: %s", absRootPath))

	return nil
}

// parseLocalZipFiles читает и распаковывает найденные zip файлы из локальной директории
func parseLocalZipFiles() error {
	return filepath.Walk(filepath.Join(config.GetString("source.local.rootPath"), "zip"), func(path string, info fs.FileInfo, err error) error {
		if err == nil && filepath.Ext(path) == ".zip" {
			slog.Info(fmt.Sprintf("Найден архив %s", info.Name()))

			files, err := zip.OpenReader(path)
			if err == nil {
				slog.Debug(fmt.Sprintf("Обработка архива %s", info.Name()))

				errCount := manageFilesInZip(files.File)

				files.Close()
				if errCount > 0 {
					slog.Warn("В архиве есть ошибочные документы. Перенос архива в дриеткорию error")
					if err := os.Rename(path, filepath.Join(config.GetString("source.local.rootPath"), "error", info.Name())); err != nil {
						slog.Error("Ошибка при переносе архива в error", slog.String("error", err.Error()))
					}
				} else {
					slog.Info(fmt.Sprintf("Архив %s успешно обработан", info.Name()))
					if err := os.Rename(path, filepath.Join(config.GetString("source.local.rootPath"), "done", info.Name())); err != nil {
						slog.Error("Ошибка при переносе архива в done", slog.String("error", err.Error()))
					}
				}
			} else {
				slog.Error("Ошибка при чтении архива", slog.String("error", err.Error()))
			}
		}

		return err
	})
}