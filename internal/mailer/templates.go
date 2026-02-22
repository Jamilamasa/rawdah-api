package mailer

import "fmt"

// BuildEmail generates a branded HTML email template.
func BuildEmail(title, body, ctaText, ctaURL, familyName string) string {
	ctaSection := ""
	if ctaText != "" && ctaURL != "" {
		ctaSection = fmt.Sprintf(`
		<div style="text-align:center; margin:32px 0;">
			<a href="%s"
			   style="background-color:#2D6A4F; color:#ffffff; text-decoration:none;
			          padding:14px 28px; border-radius:6px; font-size:16px; font-weight:600;
			          display:inline-block;">
				%s
			</a>
		</div>`, ctaURL, ctaText)
	}

	footerFamily := ""
	if familyName != "" {
		footerFamily = fmt.Sprintf(" &mdash; the %s family", familyName)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:'Segoe UI',Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:40px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0"
               style="background-color:#ffffff; border-radius:8px; overflow:hidden; box-shadow:0 2px 8px rgba(0,0,0,0.1);">
          <!-- Header -->
          <tr>
            <td style="background-color:#2D6A4F; padding:32px 40px; text-align:center;">
              <h1 style="color:#ffffff; margin:0; font-size:24px; font-weight:700;">Rawdah</h1>
              <p style="color:#B7E4C7; margin:8px 0 0; font-size:14px;">Nurturing young Muslim minds</p>
            </td>
          </tr>
          <!-- Content -->
          <tr>
            <td style="padding:40px; color:#333333;">
              <h2 style="color:#2D6A4F; margin-top:0; font-size:22px;">%s</h2>
              <div style="font-size:16px; line-height:1.6; color:#555555;">
                %s
              </div>
              %s
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="background-color:#f9f9f9; padding:24px 40px; text-align:center; border-top:1px solid #eeeeee;">
              <p style="color:#999999; font-size:13px; margin:0;">
                This email was sent by Rawdah%s.<br>
                If you have questions, contact us at
                <a href="mailto:hello@rawdah.app" style="color:#2D6A4F;">hello@rawdah.app</a>
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, title, title, body, ctaSection, footerFamily)
}
