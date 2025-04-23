package mailer

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

type MailerConfig struct {
	Host     string
	From     string
	Username string
	Password string
	Port     int
}

type Mailer struct {
	config MailerConfig
	client *smtp.Client
}

func NewMailer(config MailerConfig) (*Mailer, error) {
	mailer := &Mailer{
		config: config,
	}

	if err := mailer.connect(); err != nil {
		return nil, fmt.Errorf("failed to initialize mailer: %v", err)
	}

	return mailer, nil
}

func (m *Mailer) connect() error {
	addr := net.JoinHostPort(m.config.Host, fmt.Sprintf("%d", m.config.Port))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	client, err := smtp.NewClient(conn, m.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}

	tlsConfig := &tls.Config{
		ServerName: m.config.Host,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %v", err)
	}

	auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	m.client = client
	return nil
}

func (m *Mailer) Send(toList []string, subject string, body string) error {
	if m.client == nil {
		if err := m.connect(); err != nil {
			return err
		}
	}

	if err := m.client.Mail(m.config.From); err != nil {
		if err := m.connect(); err != nil {
			return fmt.Errorf("reconnection failed: %v", err)
		}
		if err := m.client.Mail(m.config.From); err != nil {
			return fmt.Errorf("MAIL FROM failed: %v", err)
		}
	}

	for _, recipient := range toList {
		if err := m.client.Rcpt(recipient); err != nil {
			return fmt.Errorf("RCPT TO failed: %v", err)
		}
	}

	w, err := m.client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %v", err)
	}

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		m.config.From, toList[0], subject, body)

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close failed: %v", err)
	}

	return nil
}

func (m *Mailer) Close() error {
	if m.client != nil {
		err := m.client.Quit()
		m.client = nil
		return err
	}
	return nil
}
