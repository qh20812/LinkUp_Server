package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailConfig struct {
	From     string
	Password string
}

func SendResetPasswordEmail(toEmail, userName, resetLink string) error {
	gmailUser := os.Getenv("GMAIL_USER")
	gmailPassword := os.Getenv("GMAIL_PASSWORD")

	if gmailUser == "" || gmailPassword == "" {
		return fmt.Errorf("Gmail credentials not configured")
	}

	subject := "LinkUp - Đặt lại mật khẩu"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #ddd; border-radius: 5px;">
        <h2>Đặt lại mật khẩu LinkUp</h2>
        
        <p>Xin chào <strong>%s</strong>,</p>
        
        <p>Chúng tôi nhận được yêu cầu đặt lại mật khẩu cho tài khoản của bạn. Nếu bạn không gửi yêu cầu này, vui lòng bỏ qua email này.</p>
        
        <p>Để đặt lại mật khẩu, vui lòng nhấp vào nút bên dưới:</p>
        
        <div style="margin: 30px 0;">
            <a href="%s" style="background-color: #007bff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Đặt lại mật khẩu</a>
        </div>
        
        <p style="color: #666; font-size: 12px;">
            Link này sẽ hết hạn trong 10 phút.<br>
            Hoặc sao chép link này vào trình duyệt: <br>
            <span style="word-break: break-all;">%s</span>
        </p>
        
        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
        
        <p style="color: #999; font-size: 12px;">
            Nếu bạn có thắc mắc, vui lòng liên hệ: linkup.support.qtn@gmail.com
        </p>
    </div>
</body>
</html>
    `, userName, resetLink, resetLink)

	to := []string{toEmail}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", gmailUser, gmailPassword, smtpHost)

	header := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		gmailUser, toEmail, subject)

	message := []byte(header + body)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, gmailUser, to, message)
	return err
}
