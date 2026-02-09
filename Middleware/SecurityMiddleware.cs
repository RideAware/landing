using System.Diagnostics;

namespace landing.Middleware;

public class SecurityMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ILogger<SecurityMiddleware> _logger;

    private static readonly string[] BlockedPatterns =
    {
        "python-requests", "curl", "wget", "sqlmap", "nikto",
        ".php", ".env", ".git", "wp-admin", "xmlrpc", "backup", "config"
    };

    public SecurityMiddleware(RequestDelegate next, ILogger<SecurityMiddleware> logger)
    {
        _next = next;
        _logger = logger;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var request = context.Request;
        var userAgent = request.Headers.UserAgent.ToString().ToLowerInvariant();
        var requestPath = request.Path.Value?.ToLowerInvariant() ?? "";
        var queryString = request.QueryString.Value?.ToLowerInvariant() ?? "";
        var fullUri = requestPath + queryString;

        foreach (var pattern in BlockedPatterns)
        {
            if (fullUri.Contains(pattern))
            {
                _logger.LogWarning("BLOCKED attack: {Method} {Path} from {RemoteIp}",
                    request.Method, request.Path, context.Connection.RemoteIpAddress);
                context.Response.StatusCode = StatusCodes.Status403Forbidden;
                await context.Response.WriteAsync("Access Denied");
                return;
            }

            if (userAgent.Contains(pattern))
            {
                _logger.LogWarning("BLOCKED bot: {UserAgent} from {RemoteIp}",
                    userAgent, context.Connection.RemoteIpAddress);
                context.Response.StatusCode = StatusCodes.Status403Forbidden;
                await context.Response.WriteAsync("Access Denied");
                return;
            }
        }

        var sw = Stopwatch.StartNew();
        await _next(context);
        sw.Stop();

        _logger.LogInformation("{Method} {Path} {StatusCode} {Duration}ms {RemoteIp}",
            request.Method,
            request.Path,
            context.Response.StatusCode,
            sw.ElapsedMilliseconds,
            context.Connection.RemoteIpAddress);
    }
}
