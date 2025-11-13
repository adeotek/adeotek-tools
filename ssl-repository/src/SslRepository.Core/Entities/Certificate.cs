namespace SslRepository.Core.Entities;

/// <summary>
/// Represents an SSL certificate stored in the repository
/// </summary>
public class Certificate
{
    /// <summary>
    /// Unique identifier for the certificate
    /// </summary>
    public Guid Id { get; set; }

    /// <summary>
    /// Display name for the certificate
    /// </summary>
    public string Name { get; set; } = string.Empty;

    /// <summary>
    /// Description or notes about the certificate
    /// </summary>
    public string? Description { get; set; }

    /// <summary>
    /// The group this certificate belongs to
    /// </summary>
    public Guid GroupId { get; set; }

    /// <summary>
    /// Navigation property to the group
    /// </summary>
    public Group Group { get; set; } = null!;

    /// <summary>
    /// File path where the certificate is stored on disk
    /// </summary>
    public string FilePath { get; set; } = string.Empty;

    /// <summary>
    /// Original format of the uploaded certificate (PEM, PFX, etc.)
    /// </summary>
    public CertificateFormat OriginalFormat { get; set; }

    /// <summary>
    /// Indicates whether the certificate file is encrypted on disk
    /// </summary>
    public bool IsEncrypted { get; set; }

    /// <summary>
    /// Password for PFX files or encryption key reference
    /// </summary>
    public string? Password { get; set; }

    /// <summary>
    /// Certificate subject (e.g., CN=example.com)
    /// </summary>
    public string? Subject { get; set; }

    /// <summary>
    /// Certificate issuer
    /// </summary>
    public string? Issuer { get; set; }

    /// <summary>
    /// Certificate expiration date
    /// </summary>
    public DateTime? ExpiresAt { get; set; }

    /// <summary>
    /// Indicates whether the certificate includes a private key
    /// </summary>
    public bool HasPrivateKey { get; set; }

    /// <summary>
    /// Date and time when the certificate was uploaded
    /// </summary>
    public DateTime CreatedAt { get; set; }

    /// <summary>
    /// Date and time when the certificate was last modified
    /// </summary>
    public DateTime UpdatedAt { get; set; }
}
