# Build stage
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src

# Copy solution and project files
COPY SslRepository.sln .
COPY src/SslRepository.Core/SslRepository.Core.csproj src/SslRepository.Core/
COPY src/SslRepository.Infrastructure/SslRepository.Infrastructure.csproj src/SslRepository.Infrastructure/
COPY src/SslRepository.Web/SslRepository.Web.csproj src/SslRepository.Web/
COPY tests/SslRepository.Tests/SslRepository.Tests.csproj tests/SslRepository.Tests/

# Restore dependencies
RUN dotnet restore

# Copy everything else
COPY . .

# Build the application
WORKDIR /src/src/SslRepository.Web
RUN dotnet build -c Release -o /app/build

# Publish stage
FROM build AS publish
RUN dotnet publish -c Release -o /app/publish /p:UseAppHost=false

# Runtime stage
FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS final
WORKDIR /app

# Create directories for data persistence
RUN mkdir -p /app/data /app/certificates && \
    chmod 755 /app/data /app/certificates

# Copy published application
COPY --from=publish /app/publish .

# Set environment variables
ENV ASPNETCORE_URLS=http://+:8080
ENV ConnectionStrings__DefaultConnection="Data Source=/app/data/sslrepository.db"
ENV Storage__CertificatesPath="/app/certificates"

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["dotnet", "SslRepository.Web.dll"]
