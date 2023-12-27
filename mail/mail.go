package mail

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/global"
)

type Mail struct {
	ToName   string
	ToAddr   string
	FromName string
	FromAddr string
	Subject  string
	Body     string
}

type MailMulti struct {
	ToName   string
	ToArray  []string
	ToAddr   string
	FromName string
	FromAddr string
	Subject  string
	Body     string
}

// mail creds are hard coded here in mailer func
// single addr mail
func Mailer(m Mail) error {
	//ionos mail
	/*
		smtp_user := "support@port-trak.net"
		smtp_pwd := "georgeisonlylittle"
		mailer := "smtp.ionos.com"
		smtpPort := ":587"
	*/

	smtp_user := global.EnvVars.SMTPSupportUserName
	smtp_pwd := global.EnvVars.SMTPSupportUserPW
	mailer := global.EnvVars.SMTPEndpoint
	smtpPort := ":" + global.EnvVars.SMTPPort

	msg := []byte("From: " + m.FromName + " <" + m.FromAddr + ">" + "\r\n" +
		"To: " + m.ToAddr + "\r\n" +
		"Subject: " + m.Subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		m.Body + "\r\n")
	auth := smtp.PlainAuth("", smtp_user, smtp_pwd, mailer)
	err := smtp.SendMail(mailer+smtpPort, auth, m.FromAddr, []string{m.ToAddr}, msg)
	return err
}

// multi addr mail
func MailerMulti(m MailMulti) error {
	smtp_user := global.EnvVars.SMTPSupportUserName
	smtp_pwd := global.EnvVars.SMTPSupportUserPW
	mailer := global.EnvVars.SMTPEndpoint
	smtpPort := ":" + global.EnvVars.SMTPPort
	msg := []byte("From: " + m.FromName + " <" + m.FromAddr + ">" + "\r\n" +
		"To: " + m.ToAddr + "\r\n" +
		"Subject: " + m.Subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		m.Body + "\r\n")
	auth := smtp.PlainAuth("", smtp_user, smtp_pwd, mailer)
	err := smtp.SendMail(mailer+smtpPort, auth, m.FromAddr, m.ToArray, msg)
	return err
}

// send mail with this
func SendMail(toEmail string, fromName string, fromAddr string, subj string, msgHtml string) error {

	m := Mail{ToAddr: toEmail,
		FromName: fromName,
		FromAddr: fromAddr,
		Subject:  subj,
		Body:     msgHtml}

	err := Mailer(m)
	if err != nil {
		//log.Print(err.Error())
		return err
	} else {
		return nil
	}

}

func SendMailMulti(toEmail string, toArray []string, fromName string, fromAddr string, subj string, msgHtml string) error {

	m := MailMulti{ToAddr: toEmail,
		ToArray:  toArray,
		FromName: fromName,
		FromAddr: fromAddr,
		Subject:  subj,
		Body:     msgHtml}

	err := MailerMulti(m)
	if err != nil {
		//log.Print(err.Error())
		return err
	} else {
		return nil
	}

}

// build mail string
func CraftTestEmail(msg string) (string, error) {
	var htmlStr string
	htmlStr = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml" lang="en" xml:lang="en"
	xmlns:v="urn:schemas-microsoft-com:vml"
	xmlns:o="urn:schemas-microsoft-com:office:office">
	
	<head>
	
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
	<!--[if !mso]><!-->
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<!--<![endif]-->
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Email Sample</title>
	<style type="text/css">
	html {
	width: 100%;
	}
	
	@media only screen and (min-width: 600px) {
	/* Table styles go here */
		.sizer{
			width: 600px;
			text-align: center;
			margin: auto;
		}
	}
	
	@media only screen and (max-width: 599px) {
	/* mobile styles go here */
	.sizer{
		width: 100%;
		text-align: center;
		margin: auto;
	}
	}
	</style>
	
	<!--[if gte mso 9]><xml>
	<o:OfficeDocumentSettings>
	<o:AllowPNG/>
	<o:PixelsPerInch>96</o:PixelsPerInch>
	</o:OfficeDocumentSettings>
	</xml><![endif]-->
	
	</head>
		<body>
		<!--[if mso]>
		<center>
		<table><tr><td width="580">
		<![endif]-->
			<table class="sizer" align="center" bgcolor="#ffffff" border="0" style="color: #595959;">
				<tr>
					<td>
						<div>
							<table class="full" align="center" width="100%" cellpadding="8" cellspacing="8" border="0" style="width: 100%; font-family: Helvetica, Arial, 'sans-serif'">
								<tr>
									<td width="80%" align="center">
										` + msg + `
									</td>
								</tr>
								
							</table>
						</div>
					</td>
				</tr>
			</table>
			<!--[if mso]>
			</td></tr></table>
			</center>
			<![endif]--> 
		</body>
	</html>`
	return htmlStr, nil
}

// build mail string
func CraftEmail(html string) (string, error) {
	var htmlStr string
	htmlStr = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml" lang="en" xml:lang="en"
	xmlns:v="urn:schemas-microsoft-com:vml"
	xmlns:o="urn:schemas-microsoft-com:office:office">
	
	<head>
	
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
	<!--[if !mso]><!-->
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<!--<![endif]-->
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Email Sample</title>
	<style type="text/css">
	html {
	width: 100%;
	}
	
	@media only screen and (min-width: 600px) {
	/* Table styles go here */
		.sizer{
			width: 600px;
			text-align: center;
			margin: auto;
		}
	}
	
	@media only screen and (max-width: 599px) {
	/* mobile styles go here */
	.sizer{
		width: 100%;
		text-align: center;
		margin: auto;
	}
	}
	</style>
	
	<!--[if gte mso 9]><xml>
	<o:OfficeDocumentSettings>
	<o:AllowPNG/>
	<o:PixelsPerInch>96</o:PixelsPerInch>
	</o:OfficeDocumentSettings>
	</xml><![endif]-->
	
	</head>
		<body>
		<!--[if mso]>
		<center>
		<table><tr><td width="580">
		<![endif]-->
			<table class="sizer" align="center" bgcolor="#ffffff" border="0" style="color: #595959;">
				<tr>
					<td>
						<div>
							<table class="full" align="center" width="100%" cellpadding="8" cellspacing="8" border="0" style="width: 100%; font-family: Helvetica, Arial, 'sans-serif'">
								<tr>
									<td width="80%" align="center">
										` + html + `
									</td>
								</tr>
								
							</table>
						</div>
					</td>
				</tr>
			</table>
			<!--[if mso]>
			</td></tr></table>
			</center>
			<![endif]--> 
		</body>
	</html>`
	return htmlStr, nil
}

/*
SendMAilSingle
Send an email to a single recipient
Generate the email using CraftEmail( messageHTML ) template where messageHTML will be the message content injected into the boilerplate system email template.
  - addr : string  | the reciepent address
  - emailBody : string | the output from CraftEmail( messageHTML )
  - fromNAme : string | The from name to diaply with email
  - fromAddr : string | the address to displahy as sender. For best results use the same as global.EnvVars.SMTPSupportUserName
  - subject : string | the subject field to use
*/
func SendMailSingle(addr string, emailBody string, fromName string, fromAddr string, subject string) error {
	email := addr
	sendErr := SendMail(email, fromName, fromAddr, subject, emailBody)
	if sendErr != nil {
		return sendErr
	}
	return nil
}

func SendTestEmail_handler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	email := "someRecipient@gmail.com"
	welcomeEmailStr, _ := CraftTestEmail("HI THERE!")
	err := SendMail(email, "Test Support", "support@test.com", "Test Email from Support", welcomeEmailStr)
	if err != nil {
		apierrors.HandleError(err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.SendMailError, W: &w})
		return
	}
	fmt.Fprintf(w, `mail sent`)
}
