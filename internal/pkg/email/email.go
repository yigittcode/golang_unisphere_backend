package email

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"net/smtp"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// EmailService defines the interface for email operations
type EmailService interface {
	SendVerificationEmail(toEmail, toName, token string) error
	SendWelcomeEmail(toEmail, toName string) error
}

// SMTPConfig holds configuration for SMTP server
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	FromName string
	FromEmail string
	UseTLS   bool
	BaseURL  string // Base URL for the application
}

// EmailServiceImpl implements EmailService
type EmailServiceImpl struct {
	config SMTPConfig
	logger zerolog.Logger
}

// NewEmailService creates a new EmailService
func NewEmailService(config SMTPConfig, logger zerolog.Logger) EmailService {
	return &EmailServiceImpl{
		config: config,
		logger: logger,
	}
}

// SendVerificationEmail sends an email with a verification link/token
func (s *EmailServiceImpl) SendVerificationEmail(toEmail, toName, token string) error {
	// If username or password is empty, log the email and token (for development only)
	if s.config.Username == "" || s.config.Password == "" {
		s.logger.Warn().
			Str("toEmail", toEmail).
			Str("token", token).
			Str("verificationURL", fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.config.BaseURL, token)).
			Msg("SMTP credentials not configured - verification email not sent. Use the token/URL above for testing.")
		
		// Return success for development purposes
		return nil
	}
	subject := "Verify Your Email Address - UniSphere"
	
	// Create verification URL
	verificationURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.config.BaseURL, token)
	
	body := fmt.Sprintf(`
		<html>
		<body>
			<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #333;">Welcome to UniSphere!</h2>
				<p>Hello %s,</p>
				<p>Thank you for registering with UniSphere. To complete your registration, please verify your email address by clicking the button below:</p>
				
				<div style="text-align: center; margin: 30px 0;">
					<a href="%s" style="background-color: #4a86e8; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold;">Verify Email</a>
				</div>
				
				<p>Alternatively, you can use this verification code: <strong>%s</strong></p>
				
				<p>This verification link and code will expire in 24 hours.</p>
				
				<p>If you did not register for a UniSphere account, please ignore this email.</p>
				
				<p>Best regards,<br>The UniSphere Team</p>
			</div>
		</body>
		</html>
	`, toName, verificationURL, token)
	
	return s.sendHTMLEmail(toEmail, subject, body)
}

// SendWelcomeEmail sends a welcome email to a newly verified user
func (s *EmailServiceImpl) SendWelcomeEmail(toEmail, toName string) error {
	// If username or password is empty, log the email (for development only)
	if s.config.Username == "" || s.config.Password == "" {
		s.logger.Warn().
			Str("toEmail", toEmail).
			Str("toName", toName).
			Msg("SMTP credentials not configured - welcome email not sent.")
		
		// Return success for development purposes
		return nil
	}
	subject := "Welcome to UniSphere - Your Account is Active"
	
	body := fmt.Sprintf(`
		<html>
		<body>
			<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #333;">Welcome to UniSphere!</h2>
				<p>Hello %s,</p>
				<p>Your email has been verified and your account is now active. You can now log in to access all the features of UniSphere.</p>
				
				<p>Thank you for joining our community!</p>
				
				<p>Best regards,<br>The UniSphere Team</p>
			</div>
		</body>
		</html>
	`, toName)
	
	return s.sendHTMLEmail(toEmail, subject, body)
}

// sendHTMLEmail sends an HTML email
func (s *EmailServiceImpl) sendHTMLEmail(toEmail, subject, htmlBody string) error {
	// Set up authentication information
	auth := smtp.PlainAuth(
		"",
		s.config.Username,
		s.config.Password,
		s.config.Host,
	)
	
	// Set up email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	
	// Construct message
	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + htmlBody
	
	// Connect to the server, set up a connection
	serverAddress := s.config.Host + ":" + strconv.Itoa(s.config.Port)
	
	// Use TLS if configured
	if s.config.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         s.config.Host,
		}
		
		// Connect to the SMTP server with TLS
		conn, err := tls.Dial("tcp", serverAddress, tlsConfig)
		if err != nil {
			s.logger.Error().Err(err).Str("server", serverAddress).Msg("Failed to connect to SMTP server")
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()
		
		client, err := smtp.NewClient(conn, s.config.Host)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to create SMTP client")
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()
		
		// Authenticate
		if err = client.Auth(auth); err != nil {
			s.logger.Error().Err(err).Msg("SMTP authentication failed")
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
		
		// Set the sender and recipient
		if err = client.Mail(s.config.FromEmail); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}
		if err = client.Rcpt(toEmail); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
		
		// Send the email body
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}
		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("failed to write email message: %w", err)
		}
		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close data writer: %w", err)
		}
		
		return nil
	} else {
		// Simple SMTP without TLS
		err := smtp.SendMail(
			serverAddress,
			auth,
			s.config.FromEmail,
			[]string{toEmail},
			[]byte(message),
		)
		if err != nil {
			s.logger.Error().Err(err).Str("server", serverAddress).Msg("Failed to send email")
			return fmt.Errorf("failed to send email: %w", err)
		}
		
		return nil
	}
}

// GenerateVerificationToken generates a random token for email verification
func GenerateVerificationToken() (string, error) {
	// Using crypto/rand for secure random generation
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 32)
	
	var err error
	for i := range result {
		var n *big.Int
		n, err = rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			// Continue generation with less secure method but record the error
			result[i] = chars[int(time.Now().UnixNano()%int64(len(chars)))]
		} else {
			result[i] = chars[n.Int64()]
		}
	}
	
	// If there was any error during generation, return it, but still return the token
	// since we used a fallback method
	if err != nil {
		return string(result), fmt.Errorf("secure random generation partially failed: %w", err)
	}
	
	return string(result), nil
}