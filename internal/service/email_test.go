package service

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestEmailProviderRejectsMissingConfiguration(t *testing.T) {
	provider := NewEmailProvider("", 0, "", "", "")
	if err := provider.SendResetEmail("user@example.com", "secret-token"); err == nil {
		t.Fatal("SendResetEmail() accepted missing SMTP configuration")
	}
	if err := provider.SendInvitationEmail("user@example.com", "secret-token"); err == nil {
		t.Fatal("SendInvitationEmail() accepted missing SMTP configuration")
	}
}

func TestEmailProviderDeliversThroughConfiguredSMTP(t *testing.T) {
	address, messages := startSMTPServer(t)
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	provider := NewEmailProvider(host, port, "", "", "auth@example.com", "https://app.example.com")
	if err := provider.SendResetEmail("user@example.com", "token with spaces"); err != nil {
		t.Fatalf("SendResetEmail() error = %v", err)
	}
	message := <-messages
	if !strings.Contains(message, "https://app.example.com/reset-password?token=token+with+spaces") {
		t.Fatalf("message did not contain escaped configured reset link: %q", message)
	}
	if strings.Contains(message, "To: user@example.com") {
		t.Fatal("untrusted recipient was copied into message headers")
	}
}

func TestEmailProviderReportsSMTPFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	t.Cleanup(func() { listener.Close() })
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		_, _ = writer.WriteString("220 localhost test SMTP\r\n")
		_ = writer.Flush()
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return
			}
			if strings.HasPrefix(line, "RCPT TO:") {
				_, _ = writer.WriteString("550 delivery rejected\r\n")
			} else {
				_, _ = writer.WriteString("250 ok\r\n")
			}
			_ = writer.Flush()
		}
	}()
	host, portText, _ := net.SplitHostPort(address)
	port, _ := strconv.Atoi(portText)
	provider := NewEmailProvider(host, port, "", "", "auth@example.com")
	if err := provider.SendInvitationEmail("user@example.com", "token"); err == nil {
		t.Fatal("SendInvitationEmail() hid SMTP connection failure")
	}
}

func startSMTPServer(t *testing.T) (string, <-chan string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	messages := make(chan string, 1)
	t.Cleanup(func() { listener.Close() })
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			return
		}
		defer connection.Close()
		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		writeSMTP := func(response string) { _, _ = writer.WriteString(response); _ = writer.Flush() }
		writeSMTP("220 localhost test SMTP\r\n")
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			switch {
			case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"):
				writeSMTP("250 localhost\r\n")
			case strings.HasPrefix(line, "MAIL FROM:"), strings.HasPrefix(line, "RCPT TO:"):
				writeSMTP("250 ok\r\n")
			case strings.HasPrefix(line, "DATA"):
				writeSMTP("354 end with dot\r\n")
				var message strings.Builder
				for {
					dataLine, readErr := reader.ReadString('\n')
					if readErr != nil {
						return
					}
					if dataLine == ".\r\n" {
						break
					}
					message.WriteString(dataLine)
				}
				messages <- message.String()
				writeSMTP("250 queued\r\n")
			case strings.HasPrefix(line, "QUIT"):
				writeSMTP("221 bye\r\n")
				return
			default:
				_, _ = io.WriteString(connection, "500 unsupported\r\n")
			}
		}
	}()
	return listener.Addr().String(), messages
}

func TestEmailProviderRejectsHeaderInjection(t *testing.T) {
	provider := NewEmailProvider("smtp.example.com", 587, "", "", "auth@example.com")
	if err := provider.SendResetEmail("user@example.com\r\nBcc: attacker@example.com", "secret-token"); err == nil {
		t.Fatal("SendResetEmail() accepted a recipient containing a newline")
	}
}
