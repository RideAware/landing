using System.Text.Json;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using landing.Data;
using landing.Models;
using landing.Services;

namespace landing.Pages;

[IgnoreAntiforgeryToken]
public class ContactModel : PageModel
{
    private readonly AppDbContext _db;
    private readonly EmailService _email;
    private readonly SpamDetectionService _spam;
    private readonly IConfiguration _config;
    private readonly ILogger<ContactModel> _logger;

    public ContactModel(
        AppDbContext db,
        EmailService email,
        SpamDetectionService spam,
        IConfiguration config,
        ILogger<ContactModel> logger)
    {
        _db = db;
        _email = email;
        _spam = spam;
        _config = config;
        _logger = logger;
    }

    public void OnGet()
    {
    }

    public async Task<IActionResult> OnPostAsync()
    {
        var name = Request.Form["name"].ToString().Trim();
        var email = Request.Form["email"].ToString().Trim();
        var subject = Request.Form["subject"].ToString().Trim();
        var message = Request.Form["message"].ToString().Trim();
        var subscribe = Request.Form["subscribe"].ToString() == "on";

        // Validate required fields
        if (string.IsNullOrEmpty(name) || string.IsNullOrEmpty(email) ||
            string.IsNullOrEmpty(subject) || string.IsNullOrEmpty(message))
        {
            return JsonError("All fields are required", 400);
        }

        if (!_spam.IsValidName(name))
        {
            _logger.LogWarning("Rejected submission: Invalid name format - {Name}", name);
            return JsonError("Please provide a valid name", 400);
        }

        if (!_spam.IsValidEmail(email))
        {
            _logger.LogWarning("Rejected submission: Invalid email format - {Email}", email);
            return JsonError("Please provide a valid email address", 400);
        }

        if (!_spam.IsValidSubject(subject))
        {
            _logger.LogWarning("Rejected submission: Invalid subject - {Subject}", subject);
            return JsonError("Please select a valid subject", 400);
        }

        if (message.Length < 10)
            return JsonError("Message must be at least 10 characters", 400);

        if (message.Length > 5000)
            return JsonError("Message must be less than 5000 characters", 400);

        if (!_spam.IsEnglishText(message))
        {
            _logger.LogWarning("Rejected submission: Non-English message from {Name} ({Email})", name, email);
            return JsonError("Please submit your message in English", 400);
        }

        if (_spam.IsSpamMessage(message))
        {
            _logger.LogWarning("Rejected spam submission from {Name} ({Email})", name, email);
            return JsonError("Your message was flagged as spam. Please try again with a different message.", 400);
        }

        // If subscribe checkbox is checked, add to subscribers
        if (subscribe)
        {
            try
            {
                _db.Subscribers.Add(new Subscriber { Email = email });
                await _db.SaveChangesAsync();
                _logger.LogInformation("New subscriber added: {Email}", email);
            }
            catch
            {
                _logger.LogInformation("Subscriber {Email} already exists or failed to add", email);
            }
        }

        // Send confirmation email to the user
        try
        {
            await _email.SendContactConfirmationAsync(email, name);
            _logger.LogInformation("Contact confirmation email sent to {Email}", email);
        }
        catch (Exception ex)
        {
            _logger.LogError("Failed to send contact confirmation to {Email}: {Error}", email, ex.Message);
        }

        // Send notification email to admin
        var adminEmail = _config["ADMIN_EMAIL"];
        if (!string.IsNullOrEmpty(adminEmail))
        {
            try
            {
                await _email.SendContactNotificationAsync(adminEmail, name, email, subject, message);
                _logger.LogInformation("Contact notification sent to admin: {AdminEmail}", adminEmail);
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to send contact notification to admin: {Error}", ex.Message);
            }
        }

        // Save contact message to database
        try
        {
            _db.ContactMessages.Add(new ContactMessage
            {
                Name = name,
                Email = email,
                Subject = subject,
                Message = message,
                CreatedAt = DateTime.UtcNow
            });
            await _db.SaveChangesAsync();
        }
        catch (Exception ex)
        {
            _logger.LogWarning("Failed to save contact message: {Error}", ex.Message);
        }

        _logger.LogInformation("Contact form submitted by {Name} ({Email})", name, email);

        Response.StatusCode = 201;
        return new JsonResult(new { message = "Thank you for your message. We'll get back to you soon!" });
    }

    private IActionResult JsonError(string error, int statusCode)
    {
        Response.StatusCode = statusCode;
        return new JsonResult(new { error });
    }
}
