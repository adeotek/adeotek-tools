using System.Security.Cryptography.X509Certificates;
using System.Text;
using Microsoft.Extensions.Logging;
using Org.BouncyCastle.Crypto;
using Org.BouncyCastle.OpenSsl;
using Org.BouncyCastle.Pkcs;
using Org.BouncyCastle.Security;
using Org.BouncyCastle.X509;
using SslRepository.Core.Entities;
using SslRepository.Core.Interfaces;
using X509Certificate = Org.BouncyCastle.X509.X509Certificate;

namespace SslRepository.Infrastructure.Services;

/// <summary>
/// Service for converting certificates between different formats and extracting metadata
/// </summary>
public class CertificateConversionService : ICertificateConversionService
{
    private readonly ILogger<CertificateConversionService> _logger;

    public CertificateConversionService(ILogger<CertificateConversionService> logger)
    {
        _logger = logger;
    }

    public async Task<byte[]> ConvertCertificateAsync(
        byte[] certificateData,
        CertificateFormat sourceFormat,
        CertificateFormat targetFormat,
        string? password = null,
        CancellationToken cancellationToken = default)
    {
        try
        {
            _logger.LogDebug("Converting certificate from {SourceFormat} to {TargetFormat}", sourceFormat, targetFormat);

            // If formats are the same, return as-is
            if (sourceFormat == targetFormat)
            {
                return certificateData;
            }

            // Parse the source certificate
            var (certificate, privateKey) = await ParseCertificateAsync(certificateData, sourceFormat, password, cancellationToken);

            // Convert to target format
            var result = targetFormat switch
            {
                CertificateFormat.PEM => ConvertToPem(certificate, privateKey),
                CertificateFormat.PFX => ConvertToPfx(certificate, privateKey, password ?? ""),
                CertificateFormat.DER => ConvertToDer(certificate),
                CertificateFormat.CER => ConvertToCer(certificate),
                CertificateFormat.CRT => ConvertToCrt(certificate),
                _ => throw new NotSupportedException($"Target format {targetFormat} is not supported")
            };

            _logger.LogDebug("Successfully converted certificate to {TargetFormat}", targetFormat);
            return result;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to convert certificate from {SourceFormat} to {TargetFormat}", sourceFormat, targetFormat);
            throw;
        }
    }

    public async Task<CertificateMetadata> ExtractMetadataAsync(
        byte[] certificateData,
        CertificateFormat format,
        string? password = null,
        CancellationToken cancellationToken = default)
    {
        try
        {
            var (certificate, privateKey) = await ParseCertificateAsync(certificateData, format, password, cancellationToken);

            var metadata = new CertificateMetadata(
                Subject: certificate.SubjectDN.ToString(),
                Issuer: certificate.IssuerDN.ToString(),
                ExpiresAt: certificate.NotAfter,
                ValidFrom: certificate.NotBefore,
                HasPrivateKey: privateKey != null,
                SerialNumber: certificate.SerialNumber.ToString(),
                Thumbprint: GetThumbprint(certificate)
            );

            _logger.LogDebug("Extracted metadata from certificate: Subject={Subject}", metadata.Subject);
            return metadata;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to extract metadata from certificate");
            throw;
        }
    }

    private async Task<(X509Certificate certificate, AsymmetricKeyParameter? privateKey)> ParseCertificateAsync(
        byte[] certificateData,
        CertificateFormat format,
        string? password,
        CancellationToken cancellationToken)
    {
        return await Task.Run(() =>
        {
            return format switch
            {
                CertificateFormat.PEM => ParsePem(certificateData, password),
                CertificateFormat.PFX => ParsePfx(certificateData, password),
                CertificateFormat.DER => ParseDer(certificateData),
                CertificateFormat.CER => ParseCer(certificateData),
                CertificateFormat.CRT => ParseCrt(certificateData),
                _ => throw new NotSupportedException($"Source format {format} is not supported")
            };
        }, cancellationToken);
    }

    private (X509Certificate, AsymmetricKeyParameter?) ParsePem(byte[] data, string? password)
    {
        using var reader = new StreamReader(new MemoryStream(data));
        var pemReader = new PemReader(reader, new PasswordFinder(password));

        X509Certificate? certificate = null;
        AsymmetricKeyParameter? privateKey = null;

        object? obj;
        while ((obj = pemReader.ReadObject()) != null)
        {
            if (obj is X509Certificate cert)
            {
                certificate = cert;
            }
            else if (obj is AsymmetricCipherKeyPair keyPair)
            {
                privateKey = keyPair.Private;
            }
            else if (obj is AsymmetricKeyParameter key)
            {
                privateKey = key;
            }
        }

        if (certificate == null)
        {
            throw new InvalidOperationException("No certificate found in PEM data");
        }

        return (certificate, privateKey);
    }

    private (X509Certificate, AsymmetricKeyParameter?) ParsePfx(byte[] data, string? password)
    {
        var pkcs12 = new Pkcs12Store(new MemoryStream(data), (password ?? "").ToCharArray());

        string? alias = null;
        foreach (string a in pkcs12.Aliases)
        {
            if (pkcs12.IsKeyEntry(a))
            {
                alias = a;
                break;
            }
        }

        if (alias == null)
        {
            throw new InvalidOperationException("No key entry found in PFX file");
        }

        var certEntry = pkcs12.GetCertificate(alias);
        var keyEntry = pkcs12.GetKey(alias);

        return (certEntry.Certificate, keyEntry?.Key);
    }

    private (X509Certificate, AsymmetricKeyParameter?) ParseDer(byte[] data)
    {
        var parser = new X509CertificateParser();
        var certificate = parser.ReadCertificate(data);

        if (certificate == null)
        {
            throw new InvalidOperationException("Failed to parse DER certificate");
        }

        return (certificate, null);
    }

    private (X509Certificate, AsymmetricKeyParameter?) ParseCer(byte[] data)
    {
        // CER is usually DER encoded
        return ParseDer(data);
    }

    private (X509Certificate, AsymmetricKeyParameter?) ParseCrt(byte[] data)
    {
        // CRT can be either PEM or DER, try PEM first
        try
        {
            return ParsePem(data, null);
        }
        catch
        {
            return ParseDer(data);
        }
    }

    private byte[] ConvertToPem(X509Certificate certificate, AsymmetricKeyParameter? privateKey)
    {
        using var writer = new StringWriter();
        var pemWriter = new PemWriter(writer);

        pemWriter.WriteObject(certificate);

        if (privateKey != null)
        {
            pemWriter.WriteObject(privateKey);
        }

        pemWriter.Writer.Flush();
        return Encoding.UTF8.GetBytes(writer.ToString());
    }

    private byte[] ConvertToPfx(X509Certificate certificate, AsymmetricKeyParameter? privateKey, string password)
    {
        var store = new Pkcs12StoreBuilder().Build();

        var certEntry = new X509CertificateEntry(certificate);
        store.SetCertificateEntry(certificate.SubjectDN.ToString(), certEntry);

        if (privateKey != null)
        {
            store.SetKeyEntry(
                certificate.SubjectDN.ToString(),
                new AsymmetricKeyEntry(privateKey),
                new[] { certEntry });
        }

        using var ms = new MemoryStream();
        store.Save(ms, password.ToCharArray(), new SecureRandom());
        return ms.ToArray();
    }

    private byte[] ConvertToDer(X509Certificate certificate)
    {
        return certificate.GetEncoded();
    }

    private byte[] ConvertToCer(X509Certificate certificate)
    {
        // CER is DER encoded
        return ConvertToDer(certificate);
    }

    private byte[] ConvertToCrt(X509Certificate certificate)
    {
        // CRT is typically PEM encoded
        return ConvertToPem(certificate, null);
    }

    private string GetThumbprint(X509Certificate certificate)
    {
        using var cert = new X509Certificate2(certificate.GetEncoded());
        return cert.Thumbprint;
    }

    private class PasswordFinder : IPasswordFinder
    {
        private readonly string? _password;

        public PasswordFinder(string? password)
        {
            _password = password;
        }

        public char[] GetPassword()
        {
            return (_password ?? "").ToCharArray();
        }
    }
}
