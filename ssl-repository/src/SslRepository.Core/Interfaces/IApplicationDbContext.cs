using Microsoft.EntityFrameworkCore;
using SslRepository.Core.Entities;

namespace SslRepository.Core.Interfaces;

/// <summary>
/// Database context interface for the application
/// </summary>
public interface IApplicationDbContext
{
    /// <summary>
    /// Certificates in the repository
    /// </summary>
    DbSet<Certificate> Certificates { get; }

    /// <summary>
    /// Certificate groups
    /// </summary>
    DbSet<Group> Groups { get; }

    /// <summary>
    /// API keys for authentication
    /// </summary>
    DbSet<ApiKey> ApiKeys { get; }

    /// <summary>
    /// Saves all changes made in this context to the database
    /// </summary>
    /// <param name="cancellationToken">Cancellation token</param>
    /// <returns>Number of state entries written to the database</returns>
    Task<int> SaveChangesAsync(CancellationToken cancellationToken = default);
}
