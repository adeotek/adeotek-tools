# SSL Repository

A centralized SSL certificate repository and management system with a Blazor Server frontend and REST API for certificate distribution.

## Features

- **Web-based Management Interface**: Blazor Server UI for managing certificates, groups, and API keys
- **Certificate Storage**: Store SSL certificates in various formats (PEM, PFX, DER, CER, CRT, KEY)
- **Optional Encryption**: Optionally encrypt certificates on disk for enhanced security
- **Certificate Grouping**: Organize certificates into logical groups
- **API Access Control**: Secure API access with API keys tied to specific certificate groups
- **Format Conversion**: Automatic certificate format conversion via API
- **Metadata Extraction**: Automatic extraction of certificate information (subject, issuer, expiration, etc.)
- **Multiple Deployment Options**: Deploy via Docker or IIS
- **SQLite Database**: Lightweight database for metadata and configuration

## Architecture

The application is built using .NET 10.0 and C# 14 with a clean architecture approach:

- **SslRepository.Core**: Domain models and interfaces
- **SslRepository.Infrastructure**: Data access, storage, and certificate conversion services
- **SslRepository.Web**: Blazor Server UI and Minimal API endpoints
- **SslRepository.Tests**: Unit tests with xUnit

## Prerequisites

- .NET 10.0 SDK or later (for local development)
- Docker (for containerized deployment)
- IIS with ASP.NET Core Hosting Bundle (for IIS deployment)

## Getting Started

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/adeotek/adeotek-tools.git
   cd adeotek-tools/ssl-repository
   ```

2. **Restore dependencies**
   ```bash
   dotnet restore
   ```

3. **Update configuration**
   Edit `src/SslRepository.Web/appsettings.json`:
   ```json
   {
     "ConnectionStrings": {
       "DefaultConnection": "Data Source=sslrepository.db"
     },
     "Storage": {
       "CertificatesPath": "certificates",
       "EncryptionKey": "YOUR-SECURE-ENCRYPTION-KEY-HERE",
       "EncryptByDefault": false
     }
   }
   ```

4. **Run the application**
   ```bash
   cd src/SslRepository.Web
   dotnet run
   ```

5. **Access the application**
   - Web UI: https://localhost:5001
   - API: https://localhost:5001/api/certificates

### Docker Deployment

1. **Build and run with Docker Compose**
   ```bash
   cd ssl-repository
   docker-compose up -d
   ```

2. **Access the application**
   - Web UI: http://localhost:8080
   - API: http://localhost:8080/api/certificates

3. **Configure encryption key** (recommended)
   ```bash
   ENCRYPTION_KEY="your-secure-key" docker-compose up -d
   ```

### IIS Deployment

1. **Publish the application**
   ```bash
   cd src/SslRepository.Web
   dotnet publish -c Release -o publish
   ```

2. **Copy to IIS directory**
   Copy the contents of the `publish` folder to your IIS web directory

3. **Configure IIS**
   - Create a new Application Pool with "No Managed Code"
   - Create a new website pointing to the published directory
   - Ensure the application pool identity has read/write permissions to the database and certificates directories

4. **Update web.config** if needed
   The `web.config` file is already configured for IIS deployment

## Usage

### Web Interface

1. **Create Groups**
   - Navigate to "Groups" in the sidebar
   - Click "Create Group" to organize your certificates

2. **Upload Certificates**
   - Navigate to "Certificates"
   - Click "Upload Certificate"
   - Select the certificate file, format, and group
   - Optionally enable encryption for the stored file

3. **Create API Keys**
   - Navigate to "API Keys"
   - Click "Create API Key"
   - Select which groups the key can access
   - Copy the generated API key (it won't be shown again!)

### REST API

The API supports authentication via **header** or **query parameter**:
- Header: `X-API-Key: your-api-key`
- Query parameter: `?api_key=your-api-key`

#### List Certificates
```bash
# Using header authentication
curl -H "X-API-Key: your-api-key" \
  https://your-domain/api/certificates

# Using query parameter authentication
curl "https://your-domain/api/certificates?api_key=your-api-key"
```

#### Get Certificate Metadata
```bash
# Using header authentication
curl -H "X-API-Key: your-api-key" \
  https://your-domain/api/certificates/{certificate-id}/metadata

# Using query parameter authentication
curl "https://your-domain/api/certificates/{certificate-id}/metadata?api_key=your-api-key"
```

#### Download Certificate
```bash
# Using header authentication
curl -H "X-API-Key: your-api-key" \
  https://your-domain/api/certificates/{certificate-id} \
  -o certificate.pem

# Using query parameter authentication
curl "https://your-domain/api/certificates/{certificate-id}?api_key=your-api-key" \
  -o certificate.pem
```

#### Download Certificate with Format Conversion
```bash
# Convert to PFX (header auth)
curl -H "X-API-Key: your-api-key" \
  "https://your-domain/api/certificates/{certificate-id}?format=pfx" \
  -o certificate.pfx

# Convert to DER (query param auth)
curl "https://your-domain/api/certificates/{certificate-id}?api_key=your-api-key&format=der" \
  -o certificate.der
```

## Configuration

### appsettings.json

```json
{
  "ConnectionStrings": {
    "DefaultConnection": "Data Source=sslrepository.db"
  },
  "Storage": {
    "CertificatesPath": "certificates",
    "EncryptionKey": "CHANGE-THIS-KEY-IN-PRODUCTION",
    "EncryptByDefault": false
  },
  "Logging": {
    "LogLevel": {
      "Default": "Information"
    }
  }
}
```

### Environment Variables (Docker)

- `ConnectionStrings__DefaultConnection`: Database connection string
- `Storage__CertificatesPath`: Path to store certificates
- `Storage__EncryptionKey`: Key for encrypting certificates (highly recommended to set)
- `Storage__EncryptByDefault`: Enable encryption by default (true/false)
- `ASPNETCORE_ENVIRONMENT`: Set to "Production" or "Development"

## Security Considerations

1. **Encryption Key**: Always use a strong, randomly generated encryption key in production
2. **HTTPS**: Always deploy behind HTTPS in production
3. **API Keys**: Treat API keys as secrets and rotate them regularly
4. **Access Control**: Use groups to limit API key access to specific certificates
5. **Database Security**: Ensure the SQLite database file has appropriate file permissions
6. **Certificate Storage**: The certificates directory should have restricted file permissions

## Testing

Run the unit tests:

```bash
dotnet test
```

## Contributing

Contributions are welcome! Please follow the existing code style and include tests for new features.

## License

See [LICENSE](../LICENSE) file for details.

## Support

For issues and questions, please create an issue on GitHub.
