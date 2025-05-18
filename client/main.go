package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Введите путь к файлу для отправки на сервер")

	reader := bufio.NewReader(os.Stdin)
	filePath, _ := reader.ReadString('\n')
	filePath = strings.TrimSpace(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	conn, err := net.Dial("tcp", "localhost:5555")
	if err != nil {
		fmt.Println("Ошибка подключения к серверу:", err)
		return
	}
	defer conn.Close()

	fileName := file.Name()
	_, err = conn.Write([]byte(fileName + "\n"))
	if err != nil {
		fmt.Println("Ошибка при отправке имени файла:", err)
		return
	}
	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Ошибка при отправке содержимого файла:", err)
		return
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.CloseWrite()
	}

	result, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Ошибка  при получении результата.", err)
		return
	}

	fmt.Printf("Результат анализа:\n%s\n", string(result))
}
