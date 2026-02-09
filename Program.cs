using Microsoft.EntityFrameworkCore;
using landing.Data;
using landing.Middleware;
using landing.Services;

var builder = WebApplication.CreateBuilder(args);

// Load .env file if present (for local development)
var envPath = Path.Combine(builder.Environment.ContentRootPath, ".env");
if (File.Exists(envPath))
{
    foreach (var line in File.ReadAllLines(envPath))
    {
        var trimmed = line.Trim();
        if (string.IsNullOrEmpty(trimmed) || trimmed.StartsWith('#'))
            continue;

        var idx = trimmed.IndexOf('=');
        if (idx <= 0) continue;

        var key = trimmed[..idx].Trim();
        var value = trimmed[(idx + 1)..].Trim();

        if (string.IsNullOrEmpty(Environment.GetEnvironmentVariable(key)))
            Environment.SetEnvironmentVariable(key, value);
    }
}

// Add environment variables to configuration
builder.Configuration.AddEnvironmentVariables();

// Build connection string from env vars
var pgHost = builder.Configuration["PG_HOST"] ?? "localhost";
var pgPort = builder.Configuration["PG_PORT"] ?? "5432";
var pgDatabase = builder.Configuration["PG_DATABASE"] ?? "rideaware";
var pgUser = builder.Configuration["PG_USER"] ?? "postgres";
var pgPassword = builder.Configuration["PG_PASSWORD"] ?? "";

var connectionString = $"Host={pgHost};Port={pgPort};Database={pgDatabase};Username={pgUser};Password={pgPassword}";

// Register services
builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseNpgsql(connectionString));

builder.Services.AddSingleton<EmailService>();
builder.Services.AddSingleton<SpamDetectionService>();
builder.Services.AddRazorPages();

var app = builder.Build();

// Ensure database tables exist
using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    await db.Database.EnsureCreatedAsync();
}

// Middleware pipeline
app.UseMiddleware<SecurityMiddleware>();

if (!app.Environment.IsDevelopment())
{
    app.UseExceptionHandler("/Error");
}

app.UseStaticFiles();
app.UseRouting();
app.MapRazorPages();

// Configure Kestrel to listen on the right port
var host = builder.Configuration["HOST"] ?? "0.0.0.0";
var port = builder.Configuration["PORT"] ?? "5000";

app.Run($"http://{host}:{port}");
