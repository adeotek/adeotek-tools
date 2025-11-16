using System.Security.Cryptography;
using System.Text;

using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using SslRepository.Core.Entities;
using SslRepository.Core.Interfaces;

namespace SslRepository.Infrastructure.Services;

/// <summary>
/// Service for managing certificate file storage with optional encryption
/// </summary>
public class CertificateStorageService : ICertificateStorageService
{
    private readonly ILogger<CertificateStorageService> _logger;
    private readonly StorageOptions _options;
    private readonly byte[] _encryptionKey;

    public CertificateStorageService(
        ILogger<CertificateStorageService> logger,
        IOptions<StorageOptions> options)
    {
        _logger = logger;
        _options = options.Value;

        // Ensure storage directory exists
        if (!Directory.Exists(_options.CertificatesPath))
        {
            Directory.CreateDirectory(_options.CertificatesPath);
            _logger.LogInformation("Created certificates storage directory: {Path}", _options.CertificatesPath);
        }

        // Initialize encryption key
        _encryptionKey = DeriveEncryptionKey(_options.EncryptionKey);
    }

    public async Task<string> StoreCertificateAsync(
        Guid certificateId,
        byte[] fileContent,
        CertificateFormat format,
        bool encrypt,
        CancellationToken cancellationToken = default)
    {
        try
        {
            var fileName = $"{certificateId}{GetFileExtension(format)}";
            var filePath = Path.Combine(_options.CertificatesPath, fileName);

            var dataToStore = encrypt ? EncryptData(fileContent) : fileContent;

            await File.WriteAllBytesAsync(filePath, dataToStore, cancellationToken);

            _logger.LogInformation(
                "Stored certificate {CertificateId} at {FilePath} (encrypted: {Encrypted})",
                certificateId,
                filePath,
                encrypt);

            return filePath;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to store certificate {CertificateId}", certificateId);
            throw;
        }
    }

    public async Task<byte[]> GetCertificateAsync(
        string filePath,
        bool isEncrypted,
        CancellationToken cancellationToken = default)
    {
        try
        {
            if (!File.Exists(filePath))
            {
                throw new FileNotFoundException($"Certificate file not found: {filePath}");
            }

            var fileContent = await File.ReadAllBytesAsync(filePath, cancellationToken);

            var result = isEncrypted ? DecryptData(fileContent) : fileContent;

            _logger.LogDebug("Retrieved certificate from {FilePath} (encrypted: {Encrypted})", filePath, isEncrypted);

            return result;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to retrieve certificate from {FilePath}", filePath);
            throw;
        }
    }

    public Task DeleteCertificateAsync(string filePath, CancellationToken cancellationToken = default)
    {
        try
        {
            if (File.Exists(filePath))
            {
                File.Delete(filePath);
                _logger.LogInformation("Deleted certificate file: {FilePath}", filePath);
            }
            else
            {
                _logger.LogWarning("Certificate file not found for deletion: {FilePath}", filePath);
            }

            return Task.CompletedTask;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to delete certificate file: {FilePath}", filePath);
            throw;
        }
    }

    private byte[] EncryptData(byte[] data)
    {
        using var aes = Aes.Create();
        aes.Key = _encryptionKey;
        aes.GenerateIV();

        using var encryptor = aes.CreateEncryptor();
        using var ms = new MemoryStream();

        // Write IV to the beginning of the stream
        ms.Write(aes.IV, 0, aes.IV.Length);

        using (var cs = new CryptoStream(ms, encryptor, CryptoStreamMode.Write))
        {
            cs.Write(data, 0, data.Length);
        }

        return ms.ToArray();
    }

    private byte[] DecryptData(byte[] encryptedData)
    {
        using var aes = Aes.Create();
        aes.Key = _encryptionKey;

        // Read IV from the beginning of the encrypted data
        var iv = new byte[aes.IV.Length];
        Array.Copy(encryptedData, 0, iv, 0, iv.Length);
        aes.IV = iv;

        using var decryptor = aes.CreateDecryptor();
        using var ms = new MemoryStream();
        using (var cs = new CryptoStream(ms, decryptor, CryptoStreamMode.Write))
        {
            cs.Write(encryptedData, iv.Length, encryptedData.Length - iv.Length);
        }

        return ms.ToArray();
    }

    private static byte[] DeriveEncryptionKey(string password)
    {
        // Use PBKDF2 to derive a 256-bit key from the password
        return Rfc2898DeriveBytes.Pbkdf2(
            password,
            Encoding.UTF8.GetBytes("SslRepository.Salt.v1"),
            100000,
            HashAlgorithmName.SHA256,
            32);
    }

    private static string GetFileExtension(CertificateFormat format)
    {
        return format switch
        {
            CertificateFormat.PEM => ".pem",
            CertificateFormat.PFX => ".pfx",
            CertificateFormat.DER => ".der",
            CertificateFormat.CER => ".cer",
            CertificateFormat.CRT => ".crt",
            CertificateFormat.KEY => ".key",
            _ => ".dat"
        };
    }
}

/// <summary>
/// Configuration options for certificate storage
/// </summary>
public class StorageOptions
{
    public const string SectionName = "Storage";

    /// <summary>
    /// Directory path where certificates are stored
    /// </summary>
    public string CertificatesPath { get; set; } = "certificates";

    /// <summary>
    /// Encryption key for encrypting certificates on disk
    /// </summary>
    public string EncryptionKey { get; set; } = "DefaultEncryptionKey-ChangeInProduction";

    /// <summary>
    /// Enable encryption for all stored certificates
    /// </summary>
    public bool EncryptByDefault { get; set; }
}
