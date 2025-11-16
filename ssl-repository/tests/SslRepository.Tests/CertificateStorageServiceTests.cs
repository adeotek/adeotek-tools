using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using NSubstitute;
using SslRepository.Core.Entities;
using SslRepository.Infrastructure.Services;
using Xunit;

namespace SslRepository.Tests;

public class CertificateStorageServiceTests : IDisposable
{
    private readonly string _testPath;
    private readonly CertificateStorageService _service;

    public CertificateStorageServiceTests()
    {
        _testPath = Path.Combine(Path.GetTempPath(), $"ssl-test-{Guid.NewGuid()}");
        Directory.CreateDirectory(_testPath);

        var logger = Substitute.For<ILogger<CertificateStorageService>>();
        var options = Options.Create(new StorageOptions
        {
            CertificatesPath = _testPath,
            EncryptionKey = "TestEncryptionKey123",
            EncryptByDefault = false
        });

        _service = new CertificateStorageService(logger, options);
    }

    [Fact]
    public async Task StoreCertificateAsync_ShouldStoreUnencryptedFile()
    {
        // Arrange
        var certificateId = Guid.NewGuid();
        var fileContent = System.Text.Encoding.UTF8.GetBytes("Test Certificate Content");

        // Act
        var filePath = await _service.StoreCertificateAsync(
            certificateId,
            fileContent,
            CertificateFormat.PEM,
            encrypt: false);

        // Assert
        Assert.NotNull(filePath);
        Assert.NotEmpty(filePath);
        Assert.True(File.Exists(filePath));

        var storedContent = await File.ReadAllBytesAsync(filePath);
        Assert.Equal(fileContent, storedContent);
    }

    [Fact]
    public async Task StoreCertificateAsync_ShouldStoreEncryptedFile()
    {
        // Arrange
        var certificateId = Guid.NewGuid();
        var fileContent = System.Text.Encoding.UTF8.GetBytes("Test Certificate Content");

        // Act
        var filePath = await _service.StoreCertificateAsync(
            certificateId,
            fileContent,
            CertificateFormat.PEM,
            encrypt: true);

        // Assert
        Assert.NotNull(filePath);
        Assert.NotEmpty(filePath);
        Assert.True(File.Exists(filePath));

        var storedContent = await File.ReadAllBytesAsync(filePath);
        Assert.NotEqual(fileContent, storedContent); // Content should be encrypted
    }

    [Fact]
    public async Task GetCertificateAsync_ShouldRetrieveUnencryptedFile()
    {
        // Arrange
        var certificateId = Guid.NewGuid();
        var fileContent = System.Text.Encoding.UTF8.GetBytes("Test Certificate Content");
        var filePath = await _service.StoreCertificateAsync(
            certificateId,
            fileContent,
            CertificateFormat.PEM,
            encrypt: false);

        // Act
        var retrievedContent = await _service.GetCertificateAsync(filePath, isEncrypted: false);

        // Assert
        Assert.Equal(fileContent, retrievedContent);
    }

    [Fact]
    public async Task GetCertificateAsync_ShouldRetrieveEncryptedFile()
    {
        // Arrange
        var certificateId = Guid.NewGuid();
        var fileContent = System.Text.Encoding.UTF8.GetBytes("Test Certificate Content");
        var filePath = await _service.StoreCertificateAsync(
            certificateId,
            fileContent,
            CertificateFormat.PEM,
            encrypt: true);

        // Act
        var retrievedContent = await _service.GetCertificateAsync(filePath, isEncrypted: true);

        // Assert
        Assert.Equal(fileContent, retrievedContent);
    }

    [Fact]
    public async Task DeleteCertificateAsync_ShouldRemoveFile()
    {
        // Arrange
        var certificateId = Guid.NewGuid();
        var fileContent = System.Text.Encoding.UTF8.GetBytes("Test Certificate Content");
        var filePath = await _service.StoreCertificateAsync(
            certificateId,
            fileContent,
            CertificateFormat.PEM,
            encrypt: false);

        // Act
        await _service.DeleteCertificateAsync(filePath);

        // Assert
        Assert.False(File.Exists(filePath));
    }

    public void Dispose()
    {
        if (Directory.Exists(_testPath))
        {
            Directory.Delete(_testPath, recursive: true);
        }
    }
}
