using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using landing.Data;
using landing.Models;

namespace landing.Pages;

public class NewsletterDetailModel : PageModel
{
    private readonly AppDbContext _db;

    public NewsletterDetailModel(AppDbContext db)
    {
        _db = db;
    }

    public Newsletter? NewsletterItem { get; set; }

    public async Task<IActionResult> OnGetAsync(int id)
    {
        NewsletterItem = await _db.Newsletters.FindAsync(id);

        if (NewsletterItem == null)
            return NotFound("Newsletter not found");

        return Page();
    }
}
