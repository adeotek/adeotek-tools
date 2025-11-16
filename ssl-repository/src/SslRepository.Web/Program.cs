using Microsoft.AspNetCore.Authentication;
using Microsoft.EntityFrameworkCore;
using SslRepository.Core.Interfaces;
using SslRepository.Infrastructure.Data;
using SslRepository.Infrastructure.Services;
using SslRepository.Web.Authentication;
using SslRepository.Web.Components;
using SslRepository.Web.Endpoints;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container
builder.Services.AddRazorComponents()
    .AddInteractiveServerComponents();

// Configure database
var connectionString = builder.Configuration.GetConnectionString("DefaultConnection")
    ?? "Data Source=sslrepository.db";

builder.Services.AddDbContext<ApplicationDbContext>(options =>
    options.UseSqlite(connectionString));

// Register services
builder.Services.AddScoped<IApplicationDbContext>(sp =>
    sp.GetRequiredService<ApplicationDbContext>());

builder.Services.Configure<StorageOptions>(
    builder.Configuration.GetSection(StorageOptions.SectionName));

builder.Services.AddScoped<ICertificateStorageService, CertificateStorageService>();
builder.Services.AddScoped<ICertificateConversionService, CertificateConversionService>();

// Configure API authentication
builder.Services.AddAuthentication("ApiKey")
    .AddScheme<AuthenticationSchemeOptions, ApiKeyAuthenticationHandler>("ApiKey", null);

builder.Services.AddAuthorization();

// Add OpenAPI/Swagger support
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddOpenApi();

var app = builder.Build();

// Initialize database
await DbInitializer.InitializeDatabaseAsync(app.Services);

// Configure the HTTP request pipeline
if (app.Environment.IsDevelopment())
{
    app.MapOpenApi();
}
else
{
    app.UseExceptionHandler("/Error", createScopeForErrors: true);
    app.UseHsts();
}

app.UseHttpsRedirection();
app.UseStaticFiles();
app.UseAntiforgery();

app.UseAuthentication();
app.UseAuthorization();

// Map Blazor components
app.MapRazorComponents<App>()
    .AddInteractiveServerRenderMode();

// Map API endpoints
var apiGroup = app.MapGroup("/api/certificates")
    .RequireAuthorization()
    .WithOpenApi();

apiGroup.MapCertificateEndpoints();

app.Run();
