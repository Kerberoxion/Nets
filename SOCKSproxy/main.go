package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	socks5Version     = 0x05
	socks5CommandTCP  = 0x01
	addressTypeIPv4   = 0x01
	addressTypeDomain = 0x03
)

// ProxyServer представляет конфигурацию прокси-сервера
type ProxyServer struct {
	port string
}

// NewProxyServer создает новый экземпляр прокси-сервера
func NewProxyServer(port string) *ProxyServer {
	return &ProxyServer{port: port}
}

// Run запускает прокси-сервер
func (ps *ProxyServer) Run() error {
	listener, err := net.Listen("tcp", ":"+ps.port)
	if err != nil {
		return fmt.Errorf("не удалось запустить сервер: %v", err)
	}
	defer listener.Close()

	log.Printf("Прокси-сервер слушает порт %s", ps.port)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Ошибка при принятии соединения: %v", err)
			continue
		}

		go ps.handleClient(clientConn)
	}
}

func (ps *ProxyServer) handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	// 1. Обработка handshake
	if err := ps.handleHandshake(clientConn); err != nil {
		log.Printf("Ошибка handshake: %v", err)
		return
	}

	// 2. Чтение и обработка запроса
	targetAddr, err := ps.readRequest(clientConn)
	if err != nil {
		log.Printf("Ошибка при обработке запроса: %v", err)
		return
	}

	// 3. Установка соединения с целевым адресом
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("Не удалось подключиться к целевому адресу %s: %v", targetAddr, err)
		clientConn.Write([]byte{socks5Version, 0x01, 0x00, addressTypeIPv4, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	// Уведомление клиента об успешном соединении
	clientConn.Write([]byte{socks5Version, 0x00, 0x00, addressTypeIPv4, 0, 0, 0, 0, 0, 0})

	// 4. Проксирование данных между клиентом и сервером
	ps.proxyData(clientConn, targetConn)
}

func (ps *ProxyServer) handleHandshake(clientConn net.Conn) error {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, buf); err != nil {
		return err
	}
	if buf[0] != socks5Version {
		return errors.New("неверная версия SOCKS")
	}

	numMethods := int(buf[1])
	methods := make([]byte, numMethods)
	if _, err := io.ReadFull(clientConn, methods); err != nil {
		return err
	}

	// Отвечаем "no authentication required"
	_, err := clientConn.Write([]byte{socks5Version, 0x00})
	return err
}

func (ps *ProxyServer) readRequest(clientConn net.Conn) (string, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(clientConn, buf); err != nil {
		return "", err
	}
	if buf[0] != socks5Version || buf[1] != socks5CommandTCP {
		return "", errors.New("неверный запрос")
	}

	addrType := buf[3]
	var targetAddr string
	switch addrType {
	case addressTypeIPv4:
		ip := make([]byte, 4)
		if _, err := io.ReadFull(clientConn, ip); err != nil {
			return "", err
		}
		port := make([]byte, 2)
		if _, err := io.ReadFull(clientConn, port); err != nil {
			return "", err
		}
		targetAddr = fmt.Sprintf("%s:%d", net.IP(ip).String(), binary.BigEndian.Uint16(port))
	case addressTypeDomain:
		lengthBuf := make([]byte, 1)
		if _, err := io.ReadFull(clientConn, lengthBuf); err != nil {
			return "", err
		}
		length := lengthBuf[0]
		domain := make([]byte, length)
		if _, err := io.ReadFull(clientConn, domain); err != nil {
			return "", err
		}
		port := make([]byte, 2)
		if _, err := io.ReadFull(clientConn, port); err != nil {
			return "", err
		}
		targetAddr = fmt.Sprintf("%s:%d", string(domain), binary.BigEndian.Uint16(port))
	default:
		return "", errors.New("неподдерживаемый тип адреса")
	}

	return targetAddr, nil
}

func (ps *ProxyServer) proxyData(clientConn, targetConn net.Conn) {
	done := make(chan struct{}, 2)

	// Перенаправление данных от клиента к цели
	go func() {
		io.Copy(targetConn, clientConn)
		done <- struct{}{}
	}()

	// Перенаправление данных от цели к клиенту
	go func() {
		io.Copy(clientConn, targetConn)
		done <- struct{}{}
	}()

	// Ожидаем завершения обеих горутин
	<-done
	<-done
}

func main() {
	port := "1080" // Порт по умолчанию
	proxyServer := NewProxyServer(port)

	if err := proxyServer.Run(); err != nil {
		log.Fatalf("Ошибка запуска прокси-сервера: %v", err)
	}
}
