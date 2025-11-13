using SslRepository.Core.Entities;

namespace SslRepository.Core.Interfaces;

/// <summary>
/// Service for managing certificate file storage
/// </summary>
public interface ICertificateStorageService
{
    /// <summary>
    /// Stores a certificate file on disk
    /// </summary>
    /// <param name="certificateId">Unique identifier for the certificate</param>
    /// <param name="fileContent">Certificate file content</param>
    /// <param name="format">Certificate format</param>
    /// <param name="encrypt">Whether to encrypt the file on disk</param>
    /// <param name="cancellationToken">Cancellation token</param>
    /// <returns>File path where the certificate was stored</returns>
    Task<string> StoreCertificateAsync(Guid certificateId, byte[] fileContent, CertificateFormat format, bool encrypt, CancellationToken cancellationToken = default);

    /// <summary>
    /// Retrieves a certificate file from disk
    /// </summary>
    /// <param name="filePath">Path to the certificate file</param>
    /// <param name="isEncrypted">Whether the file is encrypted</param>
    /// <param name="cancellationToken">Cancellation token</param>
    /// <returns>Certificate file content</returns>
    Task<byte[]> GetCertificateAsync(string filePath, bool isEncrypted, CancellationToken cancellationToken = default);

    /// <summary>
    /// Deletes a certificate file from disk
    /// </summary>
    /// <param name="filePath">Path to the certificate file</param>
    /// <param name="cancellationToken">Cancellation token</param>
    Task DeleteCertificateAsync(string filePath, CancellationToken cancellationToken = default);
}
