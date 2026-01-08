using SqlMigration.Models;

namespace SqlMigration.Services;

public interface IBackupService
{
    /// <summary>
    /// Creates a backup of the database
    /// </summary>
    /// <param name="connectionParameters">Database connection parameters</param>
    /// <param name="ct">Cancellation token</param>
    /// <returns>Path to the backup file</returns>
    Task<string> CreateBackupAsync(ConnectionParameters connectionParameters, CancellationToken ct = default);

    /// <summary>
    /// Restores the last backup of the database
    /// </summary>
    /// <param name="connectionParameters">Database connection parameters</param>
    /// <param name="ct">Cancellation token</param>
    Task RestoreLastBackupAsync(ConnectionParameters connectionParameters, CancellationToken ct = default);

    /// <summary>
    /// Gets the path to the last backup file
    /// </summary>
    /// <param name="connectionParameters">Database connection parameters</param>
    /// <returns>Path to the last backup file, or null if no backup exists</returns>
    string? GetLastBackupPath(ConnectionParameters connectionParameters);
}
