using System.Net.Security;
using MailKit.Net.Smtp;
using MailKit.Security;
using MimeKit;

namespace landing.Services;

public class EmailService
{
    private readonly IConfiguration _config;
    private readonly ILogger<EmailService> _logger;

    public EmailService(IConfiguration config, ILogger<EmailService> logger)
    {
        _config = config;
        _logger = logger;
    }

    public async Task SendConfirmationEmailAsync(string email, string unsubscribeLink)
    {
        var subject = "Thanks for subscribing!";
        var htmlBody = $@"
            <html>
            <body>
              <h1>Welcome to RideAware!</h1>
              <p>Thank you for subscribing to our newsletter.</p>
              <p><a href=""{System.Net.WebUtility.HtmlEncode(unsubscribeLink)}"">Unsubscribe</a></p>
            </body>
            </html>";

        await SendEmailAsync(email, subject, htmlBody);
    }

    public async Task SendContactConfirmationAsync(string email, string name)
    {
        var subject = "We received your message - RideAware";
        var escapedName = System.Net.WebUtility.HtmlEncode(name);
        var htmlBody = $@"
            <html>
            <body>
              <h2>Thank you for reaching out, {escapedName}!</h2>
              <p>We've received your message and will get back to you as soon as possible.</p>
              <p>In the meantime, feel free to check out more about RideAware on our website.</p>
              <p>Best regards,<br>The RideAware Team</p>
            </body>
            </html>";

        await SendEmailAsync(email, subject, htmlBody);
    }

    public async Task SendContactNotificationAsync(string adminEmail, string name, string email, string contactSubject, string message)
    {
        var emailSubject = $"New contact message from {name}";
        var escapedName = System.Net.WebUtility.HtmlEncode(name);
        var escapedEmail = System.Net.WebUtility.HtmlEncode(email);
        var escapedSubject = System.Net.WebUtility.HtmlEncode(contactSubject);
        var escapedMessage = System.Net.WebUtility.HtmlEncode(message).Replace("\n", "<br>");

        var htmlBody = $@"
            <html>
            <body>
              <h3>New Contact Message</h3>
              <p><strong>From:</strong> {escapedName} ({escapedEmail})</p>
              <p><strong>Subject:</strong> {escapedSubject}</p>
              <h4>Message:</h4>
              <p>{escapedMessage}</p>
            </body>
            </html>";

        await SendEmailAsync(adminEmail, emailSubject, htmlBody);
    }

    private async Task SendEmailAsync(string toEmail, string subject, string htmlBody)
    {
        var smtpHost = _config["SMTP_SERVER"] ?? throw new InvalidOperationException("SMTP_SERVER not configured");
        var smtpPort = int.Parse(_config["SMTP_PORT"] ?? "465");
        var smtpUser = _config["SMTP_USER"] ?? throw new InvalidOperationException("SMTP_USER not configured");
        var smtpPass = _config["SMTP_PASSWORD"] ?? throw new InvalidOperationException("SMTP_PASSWORD not configured");

        var message = new MimeMessage();
        message.From.Add(MailboxAddress.Parse(smtpUser));
        message.To.Add(MailboxAddress.Parse(toEmail));
        message.Subject = subject;
        message.Body = new TextPart("html") { Text = htmlBody };

        using var client = new SmtpClient();

        // Accept the server's certificate (matching Go's TLS behavior)
        client.ServerCertificateValidationCallback = (sender, certificate, chain, errors) =>
            errors == SslPolicyErrors.None || errors == SslPolicyErrors.RemoteCertificateChainErrors;

        // Port 465 uses direct SSL/TLS (SslOnConnect)
        await client.ConnectAsync(smtpHost, smtpPort, SecureSocketOptions.SslOnConnect);
        await client.AuthenticateAsync(smtpUser, smtpPass);
        await client.SendAsync(message);
        await client.DisconnectAsync(true);
    }
}
