using SslRepository.Web.Authentication;
using Xunit;

namespace SslRepository.Tests;

public class ApiKeyHelperTests
{
    [Fact]
    public void GenerateApiKey_ShouldReturnNonEmptyString()
    {
        // Act
        var apiKey = ApiKeyHelper.GenerateApiKey();

        // Assert
        Assert.NotNull(apiKey);
        Assert.NotEmpty(apiKey);
        Assert.True(apiKey.Length > 20);
    }

    [Fact]
    public void GenerateApiKey_ShouldReturnUniqueKeys()
    {
        // Act
        var apiKey1 = ApiKeyHelper.GenerateApiKey();
        var apiKey2 = ApiKeyHelper.GenerateApiKey();

        // Assert
        Assert.NotEqual(apiKey1, apiKey2);
    }

    [Fact]
    public void HashApiKey_ShouldReturnConsistentHash()
    {
        // Arrange
        var apiKey = "test-api-key-123";

        // Act
        var hash1 = ApiKeyHelper.HashApiKey(apiKey);
        var hash2 = ApiKeyHelper.HashApiKey(apiKey);

        // Assert
        Assert.Equal(hash1, hash2);
    }

    [Fact]
    public void HashApiKey_ShouldReturnDifferentHashesForDifferentKeys()
    {
        // Arrange
        var apiKey1 = "test-api-key-123";
        var apiKey2 = "test-api-key-456";

        // Act
        var hash1 = ApiKeyHelper.HashApiKey(apiKey1);
        var hash2 = ApiKeyHelper.HashApiKey(apiKey2);

        // Assert
        Assert.NotEqual(hash1, hash2);
    }
}
