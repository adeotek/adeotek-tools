using SslRepository.Core.Entities;

namespace SslRepository.Core.Interfaces;

/// <summary>
/// Service for converting certificates between different formats
/// </summary>
public interface ICertificateConversionService
{
    /// <summary>
    /// Converts a certificate from one format to another
    /// </summary>
    /// <param name="certificateData">Source certificate data</param>
    /// <param name="sourceFormat">Source format</param>
    /// <param name="targetFormat">Target format</param>
    /// <param name="password">Password for PFX files</param>
    /// <param name="cancellationToken">Cancellation token</param>
    /// <returns>Converted certificate data</returns>
    Task<byte[]> ConvertCertificateAsync(byte[] certificateData, CertificateFormat sourceFormat, CertificateFormat targetFormat, string? password = null, CancellationToken cancellationToken = default);

    /// <summary>
    /// Extracts certificate metadata
    /// </summary>
    /// <param name="certificateData">Certificate data</param>
    /// <param name="format">Certificate format</param>
    /// <param name="password">Password for PFX files</param>
    /// <param name="cancellationToken">Cancellation token</param>
    /// <returns>Certificate metadata</returns>
    Task<CertificateMetadata> ExtractMetadataAsync(byte[] certificateData, CertificateFormat format, string? password = null, CancellationToken cancellationToken = default);
}

/// <summary>
/// Certificate metadata extracted from a certificate file
/// </summary>
public record CertificateMetadata(
    string Subject,
    string Issuer,
    DateTime ExpiresAt,
    DateTime ValidFrom,
    bool HasPrivateKey,
    string? SerialNumber = null,
    string? Thumbprint = null
);
