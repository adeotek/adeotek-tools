using System.Security.Claims;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;
using SslRepository.Core.Entities;
using SslRepository.Core.Interfaces;
using SslRepository.Infrastructure.Data;
using SslRepository.Web.Models;

namespace SslRepository.Web.Endpoints;

/// <summary>
/// Minimal API endpoints for certificate operations
/// </summary>
public static class CertificateEndpoints
{
    public static void MapCertificateEndpoints(this RouteGroupBuilder group)
    {
        group.MapGet("/", ListCertificates)
            .WithName("ListCertificates")
            .WithDescription("Lists all certificates accessible by the authenticated API key")
            .Produces<IEnumerable<CertificateDto>>();

        group.MapGet("/{id:guid}", GetCertificate)
            .WithName("GetCertificate")
            .WithDescription("Gets a specific certificate by ID in the requested format")
            .Produces<FileContentHttpResult>()
            .Produces(StatusCodes.Status404NotFound);

        group.MapGet("/{id:guid}/metadata", GetCertificateMetadata)
            .WithName("GetCertificateMetadata")
            .WithDescription("Gets certificate metadata")
            .Produces<CertificateDto>()
            .Produces(StatusCodes.Status404NotFound);
    }

    private static async Task<IResult> ListCertificates(
        ClaimsPrincipal user,
        ApplicationDbContext dbContext)
    {
        var groupIds = GetAccessibleGroupIds(user);

        var certificates = await dbContext.Certificates
            .Where(c => groupIds.Contains(c.GroupId))
            .Include(c => c.Group)
            .Select(c => new CertificateDto(
                c.Id,
                c.Name,
                c.Description,
                c.Group.Name,
                c.Subject,
                c.Issuer,
                c.ExpiresAt,
                c.HasPrivateKey,
                c.OriginalFormat.ToString()
            ))
            .ToListAsync();

        return Results.Ok(certificates.AsEnumerable());
    }

    private static async Task<IResult> GetCertificate(
        Guid id,
        string? format,
        ClaimsPrincipal user,
        ApplicationDbContext dbContext,
        ICertificateStorageService storageService,
        ICertificateConversionService conversionService,
        ILoggerFactory loggerFactory)
    {
        var groupIds = GetAccessibleGroupIds(user);

        var certificate = await dbContext.Certificates
            .FirstOrDefaultAsync(c => c.Id == id && groupIds.Contains(c.GroupId));

        if (certificate is null)
        {
            return Results.NotFound(new { error = "Certificate not found or access denied" });
        }

        try
        {
            var logger = loggerFactory.CreateLogger("CertificateEndpoints");

            // Retrieve the certificate file
            var certificateData = await storageService.GetCertificateAsync(
                certificate.FilePath,
                certificate.IsEncrypted);

            // Convert format if requested
            CertificateFormat targetFormat = certificate.OriginalFormat;
            if (!string.IsNullOrEmpty(format) &&
                Enum.TryParse<CertificateFormat>(format, true, out var parsedFormat))
            {
                targetFormat = parsedFormat;
                if (targetFormat != certificate.OriginalFormat)
                {
                    certificateData = await conversionService.ConvertCertificateAsync(
                        certificateData,
                        certificate.OriginalFormat,
                        targetFormat,
                        certificate.Password);

                    logger.LogInformation(
                        "Converted certificate {CertificateId} from {SourceFormat} to {TargetFormat}",
                        id,
                        certificate.OriginalFormat,
                        targetFormat);
                }
            }

            var contentType = GetContentType(targetFormat);
            var fileName = $"{certificate.Name}{GetFileExtension(targetFormat)}";

            return Results.File(certificateData, contentType, fileName);
        }
        catch (Exception ex)
        {
            var logger = loggerFactory.CreateLogger("CertificateEndpoints");
            logger.LogError(ex, "Failed to retrieve certificate {CertificateId}", id);
            return Results.Problem("Failed to retrieve certificate", statusCode: 500);
        }
    }

    private static async Task<IResult> GetCertificateMetadata(
        Guid id,
        ClaimsPrincipal user,
        ApplicationDbContext dbContext)
    {
        var groupIds = GetAccessibleGroupIds(user);

        var certificate = await dbContext.Certificates
            .Include(c => c.Group)
            .FirstOrDefaultAsync(c => c.Id == id && groupIds.Contains(c.GroupId));

        if (certificate is null)
        {
            return Results.NotFound(new { error = "Certificate not found or access denied" });
        }

        var dto = new CertificateDto(
            certificate.Id,
            certificate.Name,
            certificate.Description,
            certificate.Group.Name,
            certificate.Subject,
            certificate.Issuer,
            certificate.ExpiresAt,
            certificate.HasPrivateKey,
            certificate.OriginalFormat.ToString()
        );

        return Results.Ok(dto);
    }

    private static List<Guid> GetAccessibleGroupIds(ClaimsPrincipal user)
    {
        return user.Claims
            .Where(c => c.Type == "Group")
            .Select(c => Guid.Parse(c.Value))
            .ToList();
    }

    private static string GetContentType(CertificateFormat format) => format switch
    {
        CertificateFormat.PEM => "application/x-pem-file",
        CertificateFormat.PFX => "application/x-pkcs12",
        CertificateFormat.DER => "application/x-x509-ca-cert",
        CertificateFormat.CER => "application/x-x509-ca-cert",
        CertificateFormat.CRT => "application/x-x509-ca-cert",
        CertificateFormat.KEY => "application/pkcs8",
        _ => "application/octet-stream"
    };

    private static string GetFileExtension(CertificateFormat format) => format switch
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
