package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex

func main() {
	startServer()

}

func startServer() {
	ln, err := net.Listen("tcp", ":5555")

	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}

	defer ln.Close()

	fmt.Println("Сервер запущен на порту 5555")

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		fmt.Println("Ошибка создания папки uploads:", err)
	}

	for {
		log.Println("Ожидание подключения...")

		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка при принятии подключения: ", err)
		} else {
			log.Printf("Подключение принято: %s\n", conn.RemoteAddr())
			go handleConnection(conn)
		}
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)

	fileName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(conn, "Ошибка при получении имени файла: %v\n", err)
		return
	}

	fileName = filepath.Base(strings.TrimSpace(fileName))

	timestamp := time.Now().Unix()
	uniqueFileName := fmt.Sprintf("%d_%s", timestamp, fileName)

	filePath := filepath.Join("uploads", uniqueFileName)

	log.Printf("Начало обработки файла: %s", fileName)

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(conn, "Ошибка при сохранении файла: %v\n", err)
		return
	}
	defer file.Close()
	_, err = io.Copy(file, reader)

	if err != nil && err != io.EOF {
		fmt.Fprintf(conn, "Ошибка при получении содержимого файла: %v\n", err)
		return
	}

	lines, words, chars, err := analyzeFile(filePath)
	if err != nil {
		fmt.Fprintf(conn, "Ошибка при анализе файла: %v\n", err)
		return
	}

	result := fmt.Sprintf("Имя файла: %s\nСтрок: %d, Слов: %d, Символов: %d\n", fileName, lines, words, chars)
	err = saveAnalysisResults(result)
	if err != nil {
		fmt.Println("Ошибка при сохранении результатов", err)
	}

	log.Printf("Завершена обработка файла: %s", fileName)

	conn.Write([]byte(result))
}

func analyzeFile(fileName string) (int, int, int, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()

	var lines, words, chars int

	wordRegex := regexp.MustCompile(`\b[\wа-яА-ЯёЁ]+\b`)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		lines++

		chars += len(line)

		matches := wordRegex.FindAllString(line, -1)
		words += len(matches)
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, 0, err
	}

	return lines, words, chars, nil
}

func saveAnalysisResults(result string) error {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.OpenFile("analysis_result.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Ошибка при открытии файла: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(result)
	if err != nil {
		return fmt.Errorf("Ошибка при записи в файл: %v", err)
	}

	return nil
}
