using FluentAssertions;
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
        apiKey.Should().NotBeNullOrEmpty();
        apiKey.Length.Should().BeGreaterThan(20);
    }

    [Fact]
    public void GenerateApiKey_ShouldReturnUniqueKeys()
    {
        // Act
        var apiKey1 = ApiKeyHelper.GenerateApiKey();
        var apiKey2 = ApiKeyHelper.GenerateApiKey();

        // Assert
        apiKey1.Should().NotBe(apiKey2);
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
        hash1.Should().Be(hash2);
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
        hash1.Should().NotBe(hash2);
    }
}
