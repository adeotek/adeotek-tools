namespace SslRepository.Core.Entities;

/// <summary>
/// Supported certificate formats
/// </summary>
public enum CertificateFormat
{
    /// <summary>
    /// PEM format (Privacy-Enhanced Mail) - Base64 encoded
    /// </summary>
    PEM = 0,

    /// <summary>
    /// PFX/PKCS#12 format - Binary format with private key
    /// </summary>
    PFX = 1,

    /// <summary>
    /// DER format - Binary encoded certificate
    /// </summary>
    DER = 2,

    /// <summary>
    /// CER format - Certificate file
    /// </summary>
    CER = 3,

    /// <summary>
    /// CRT format - Certificate file (similar to CER)
    /// </summary>
    CRT = 4,

    /// <summary>
    /// KEY format - Private key file
    /// </summary>
    KEY = 5
}
