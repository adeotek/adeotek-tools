namespace SslRepository.Web.Models;

/// <summary>
/// DTO for certificate information
/// </summary>
public record CertificateDto(
    Guid Id,
    string Name,
    string? Description,
    string GroupName,
    string? Subject,
    string? Issuer,
    DateTime? ExpiresAt,
    bool HasPrivateKey,
    string OriginalFormat
);
