using Microsoft.EntityFrameworkCore;
using landing.Models;

namespace landing.Data;

public class AppDbContext : DbContext
{
    public AppDbContext(DbContextOptions<AppDbContext> options) : base(options) { }

    public DbSet<Subscriber> Subscribers => Set<Subscriber>();
    public DbSet<Newsletter> Newsletters => Set<Newsletter>();
    public DbSet<ContactMessage> ContactMessages => Set<ContactMessage>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Subscriber>(entity =>
        {
            entity.HasIndex(e => e.Email).IsUnique();
        });
    }
}
