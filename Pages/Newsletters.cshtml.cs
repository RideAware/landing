using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.EntityFrameworkCore;
using landing.Data;
using landing.Models;

namespace landing.Pages;

public class NewslettersModel : PageModel
{
    private readonly AppDbContext _db;

    public NewslettersModel(AppDbContext db)
    {
        _db = db;
    }

    public List<Newsletter> NewslettersList { get; set; } = new();

    public async Task OnGetAsync()
    {
        NewslettersList = await _db.Newsletters
            .OrderByDescending(n => n.SentAt)
            .ToListAsync();
    }
}
