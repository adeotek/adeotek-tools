using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using SslRepository.Core.Entities;
using SslRepository.Core.Interfaces;
using SslRepository.Infrastructure.Data;

namespace SslRepository.Web.Controllers;

/// <summary>
/// API controller for certificate operations
/// </summary>
[ApiController]
[Route("api/certificates")]
[Authorize(AuthenticationSchemes = "ApiKey")]
public class CertificatesApiController : ControllerBase
{
    private readonly ApplicationDbContext _dbContext;
    private readonly ICertificateStorageService _storageService;
    private readonly ICertificateConversionService _conversionService;
    private readonly ILogger<CertificatesApiController> _logger;

    public CertificatesApiController(
        ApplicationDbContext dbContext,
        ICertificateStorageService storageService,
        ICertificateConversionService conversionService,
        ILogger<CertificatesApiController> logger)
    {
        _dbContext = dbContext;
        _storageService = storageService;
        _conversionService = conversionService;
        _logger = logger;
    }

    /// <summary>
    /// Lists all certificates accessible by the authenticated API key
    /// </summary>
    [HttpGet]
    public async Task<ActionResult<IEnumerable<CertificateDto>>> ListCertificates()
    {
        var groupIds = GetAccessibleGroupIds();

        var certificates = await _dbContext.Certificates
            .Where(c => groupIds.Contains(c.GroupId))
            .Include(c => c.Group)
            .Select(c => new CertificateDto
            {
                Id = c.Id,
                Name = c.Name,
                Description = c.Description,
                GroupName = c.Group.Name,
                Subject = c.Subject,
                Issuer = c.Issuer,
                ExpiresAt = c.ExpiresAt,
                HasPrivateKey = c.HasPrivateKey,
                OriginalFormat = c.OriginalFormat.ToString()
            })
            .ToListAsync();

        return Ok(certificates);
    }

    /// <summary>
    /// Gets a specific certificate by ID in the requested format
    /// </summary>
    [HttpGet("{id}")]
    public async Task<ActionResult> GetCertificate(
        Guid id,
        [FromQuery] string? format = null)
    {
        var groupIds = GetAccessibleGroupIds();

        var certificate = await _dbContext.Certificates
            .FirstOrDefaultAsync(c => c.Id == id && groupIds.Contains(c.GroupId));

        if (certificate == null)
        {
            return NotFound(new { error = "Certificate not found or access denied" });
        }

        try
        {
            // Retrieve the certificate file
            var certificateData = await _storageService.GetCertificateAsync(
                certificate.FilePath,
                certificate.IsEncrypted);

            // Convert format if requested
            if (!string.IsNullOrEmpty(format) &&
                Enum.TryParse<CertificateFormat>(format, true, out var targetFormat) &&
                targetFormat != certificate.OriginalFormat)
            {
                certificateData = await _conversionService.ConvertCertificateAsync(
                    certificateData,
                    certificate.OriginalFormat,
                    targetFormat,
                    certificate.Password);

                _logger.LogInformation(
                    "Converted certificate {CertificateId} from {SourceFormat} to {TargetFormat}",
                    id,
                    certificate.OriginalFormat,
                    targetFormat);
            }

            var contentType = GetContentType(format != null && Enum.TryParse<CertificateFormat>(format, true, out var fmt) ? fmt : certificate.OriginalFormat);
            var fileName = $"{certificate.Name}{GetFileExtension(format != null && Enum.TryParse<CertificateFormat>(format, true, out var fmt2) ? fmt2 : certificate.OriginalFormat)}";

            return File(certificateData, contentType, fileName);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to retrieve certificate {CertificateId}", id);
            return StatusCode(500, new { error = "Failed to retrieve certificate" });
        }
    }

    /// <summary>
    /// Gets certificate metadata
    /// </summary>
    [HttpGet("{id}/metadata")]
    public async Task<ActionResult<CertificateDto>> GetCertificateMetadata(Guid id)
    {
        var groupIds = GetAccessibleGroupIds();

        var certificate = await _dbContext.Certificates
            .Include(c => c.Group)
            .FirstOrDefaultAsync(c => c.Id == id && groupIds.Contains(c.GroupId));

        if (certificate == null)
        {
            return NotFound(new { error = "Certificate not found or access denied" });
        }

        var dto = new CertificateDto
        {
            Id = certificate.Id,
            Name = certificate.Name,
            Description = certificate.Description,
            GroupName = certificate.Group.Name,
            Subject = certificate.Subject,
            Issuer = certificate.Issuer,
            ExpiresAt = certificate.ExpiresAt,
            HasPrivateKey = certificate.HasPrivateKey,
            OriginalFormat = certificate.OriginalFormat.ToString()
        };

        return Ok(dto);
    }

    private List<Guid> GetAccessibleGroupIds()
    {
        var groupClaims = User.Claims
            .Where(c => c.Type == "Group")
            .Select(c => Guid.Parse(c.Value))
            .ToList();

        return groupClaims;
    }

    private static string GetContentType(CertificateFormat format)
    {
        return format switch
        {
            CertificateFormat.PEM => "application/x-pem-file",
            CertificateFormat.PFX => "application/x-pkcs12",
            CertificateFormat.DER => "application/x-x509-ca-cert",
            CertificateFormat.CER => "application/x-x509-ca-cert",
            CertificateFormat.CRT => "application/x-x509-ca-cert",
            CertificateFormat.KEY => "application/pkcs8",
            _ => "application/octet-stream"
        };
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
/// DTO for certificate information
/// </summary>
public class CertificateDto
{
    public Guid Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string? Description { get; set; }
    public string GroupName { get; set; } = string.Empty;
    public string? Subject { get; set; }
    public string? Issuer { get; set; }
    public DateTime? ExpiresAt { get; set; }
    public bool HasPrivateKey { get; set; }
    public string OriginalFormat { get; set; } = string.Empty;
}
