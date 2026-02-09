using System.Text.Json;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using landing.Data;
using landing.Models;
using landing.Services;

namespace landing.Pages;

[IgnoreAntiforgeryToken]
public class SubscribeModel : PageModel
{
    private readonly AppDbContext _db;
    private readonly EmailService _email;
    private readonly ILogger<SubscribeModel> _logger;

    public SubscribeModel(AppDbContext db, EmailService email, ILogger<SubscribeModel> logger)
    {
        _db = db;
        _email = email;
        _logger = logger;
    }

    public IActionResult OnGet() => NotFound();

    public async Task<IActionResult> OnPostAsync()
    {
        string? emailAddress = null;

        // Try to read JSON body
        try
        {
            using var reader = new StreamReader(Request.Body);
            var body = await reader.ReadToEndAsync();
            var json = JsonSerializer.Deserialize<JsonElement>(body);
            emailAddress = json.GetProperty("email").GetString();
        }
        catch
        {
            return BadRequest(new { error = "Invalid request" });
        }

        if (string.IsNullOrWhiteSpace(emailAddress))
        {
            return BadRequest(new { error = "Email is required" });
        }

        try
        {
            _db.Subscribers.Add(new Subscriber { Email = emailAddress });
            await _db.SaveChangesAsync();
        }
        catch
        {
            return BadRequest(new { error = "Email already exists" });
        }

        // Build unsubscribe link
        var scheme = Request.Scheme;
        var host = Request.Host.Value;
        var unsubscribeLink = $"{scheme}://{host}/unsubscribe?email={emailAddress}";

        try
        {
            await _email.SendConfirmationEmailAsync(emailAddress, unsubscribeLink);
            _logger.LogInformation("Confirmation email sent to {Email}", emailAddress);
        }
        catch (Exception ex)
        {
            _logger.LogError("Failed to send confirmation email to {Email}: {Error}", emailAddress, ex.Message);
        }

        Response.StatusCode = 201;
        return new JsonResult(new { message = "Email has been added" });
    }
}
