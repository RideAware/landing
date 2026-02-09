using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.EntityFrameworkCore;
using landing.Data;

namespace landing.Pages;

public class UnsubscribeModel : PageModel
{
    private readonly AppDbContext _db;
    private readonly ILogger<UnsubscribeModel> _logger;

    public UnsubscribeModel(AppDbContext db, ILogger<UnsubscribeModel> logger)
    {
        _db = db;
        _logger = logger;
    }

    public string ResultMessage { get; set; } = string.Empty;

    public async Task<IActionResult> OnGetAsync([FromQuery] string? email)
    {
        if (string.IsNullOrWhiteSpace(email))
        {
            Response.StatusCode = 400;
            ResultMessage = "No email specified";
            return Page();
        }

        var subscriber = await _db.Subscribers.FirstOrDefaultAsync(s => s.Email == email);
        if (subscriber == null)
        {
            Response.StatusCode = 400;
            ResultMessage = $"Email {email} was not found or already unsubscribed";
            return Page();
        }

        _db.Subscribers.Remove(subscriber);
        await _db.SaveChangesAsync();

        _logger.LogInformation("Unsubscribed {Email}", email);
        ResultMessage = $"The email {email} has been unsubscribed.";
        return Page();
    }
}
