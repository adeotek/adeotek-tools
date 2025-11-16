using System.Security.Claims;
using System.Text;
using System.Text.Encodings.Web;
using Microsoft.AspNetCore.Authentication;
using Microsoft.AspNetCore.Http;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using NSubstitute;
using SslRepository.Core.Entities;
using SslRepository.Infrastructure.Data;
using SslRepository.Web.Authentication;
using Xunit;

namespace SslRepository.Tests;

public class ApiKeyAuthenticationTests : IDisposable
{
    private readonly ApplicationDbContext _dbContext;
    private readonly ApiKeyAuthenticationHandler _handler;
    private readonly DefaultHttpContext _httpContext;

    public ApiKeyAuthenticationTests()
    {
        // Setup in-memory database
        var options = new DbContextOptionsBuilder<ApplicationDbContext>()
            .UseInMemoryDatabase(databaseName: $"TestDb_{Guid.NewGuid()}")
            .Options;

        _dbContext = new ApplicationDbContext(options);

        // Create test data
        SeedTestData();

        // Setup authentication handler
        var authOptions = Substitute.For<IOptionsMonitor<AuthenticationSchemeOptions>>();
        authOptions.Get(Arg.Any<string>()).Returns(new AuthenticationSchemeOptions());

        var loggerFactory = Substitute.For<ILoggerFactory>();
        var logger = Substitute.For<ILogger<ApiKeyAuthenticationHandler>>();
        loggerFactory.CreateLogger<ApiKeyAuthenticationHandler>().Returns(logger);

        var encoder = UrlEncoder.Default;

        _handler = new ApiKeyAuthenticationHandler(authOptions, loggerFactory, encoder, _dbContext);

        // Setup HTTP context
        _httpContext = new DefaultHttpContext();
        var scheme = new AuthenticationScheme("ApiKey", "ApiKey", typeof(ApiKeyAuthenticationHandler));
        _handler.InitializeAsync(scheme, _httpContext).Wait();
    }

    private void SeedTestData()
    {
        var group = new Group
        {
            Id = Guid.NewGuid(),
            Name = "Test Group"
        };

        var apiKey = new ApiKey
        {
            Id = Guid.NewGuid(),
            Name = "Test Key",
            KeyHash = ApiKeyHelper.HashApiKey("test-api-key-123"),
            IsActive = true,
            Groups = [group]
        };

        _dbContext.Groups.Add(group);
        _dbContext.ApiKeys.Add(apiKey);
        _dbContext.SaveChanges();
    }

    [Fact]
    public async Task AuthenticateAsync_WithValidHeaderApiKey_ShouldSucceed()
    {
        // Arrange
        _httpContext.Request.Headers["X-API-Key"] = "test-api-key-123";

        // Act
        var result = await _handler.AuthenticateAsync();

        // Assert
        Assert.True(result.Succeeded);
        Assert.NotNull(result.Principal);
        Assert.Equal("Test Key", result.Principal.Identity?.Name);
    }

    [Fact]
    public async Task AuthenticateAsync_WithValidQueryParameterApiKey_ShouldSucceed()
    {
        // Arrange
        _httpContext.Request.QueryString = new QueryString("?api_key=test-api-key-123");

        // Act
        var result = await _handler.AuthenticateAsync();

        // Assert
        Assert.True(result.Succeeded);
        Assert.NotNull(result.Principal);
        Assert.Equal("Test Key", result.Principal.Identity?.Name);
    }

    [Fact]
    public async Task AuthenticateAsync_WithInvalidApiKey_ShouldFail()
    {
        // Arrange
        _httpContext.Request.Headers["X-API-Key"] = "invalid-api-key";

        // Act
        var result = await _handler.AuthenticateAsync();

        // Assert
        Assert.False(result.Succeeded);
    }

    [Fact]
    public async Task AuthenticateAsync_WithNoApiKey_ShouldReturnNoResult()
    {
        // Arrange
        // No API key set

        // Act
        var result = await _handler.AuthenticateAsync();

        // Assert
        Assert.False(result.Succeeded);
        Assert.True(result.None);
    }

    [Fact]
    public async Task AuthenticateAsync_HeaderTakesPrecedenceOverQueryParam()
    {
        // Arrange
        _httpContext.Request.Headers["X-API-Key"] = "test-api-key-123";
        _httpContext.Request.QueryString = new QueryString("?api_key=invalid-key");

        // Act
        var result = await _handler.AuthenticateAsync();

        // Assert - Should succeed because header key is valid
        Assert.True(result.Succeeded);
    }

    [Fact]
    public async Task AuthenticateAsync_WithValidApiKey_ShouldIncludeGroupClaims()
    {
        // Arrange
        _httpContext.Request.Headers["X-API-Key"] = "test-api-key-123";

        // Act
        var result = await _handler.AuthenticateAsync();

        // Assert
        Assert.True(result.Succeeded);
        Assert.NotNull(result.Principal);

        var groupClaims = result.Principal.Claims.Where(c => c.Type == "Group").ToList();
        Assert.Single(groupClaims);
    }

    public void Dispose()
    {
        _dbContext.Database.EnsureDeleted();
        _dbContext.Dispose();
    }
}
