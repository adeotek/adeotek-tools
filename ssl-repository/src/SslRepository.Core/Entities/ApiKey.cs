namespace SslRepository.Core.Entities;

/// <summary>
/// Represents an API key for authenticating external clients
/// </summary>
public class ApiKey
{
    /// <summary>
    /// Unique identifier for the API key
    /// </summary>
    public Guid Id { get; set; }

    /// <summary>
    /// Display name for the API key
    /// </summary>
    public string Name { get; set; } = string.Empty;

    /// <summary>
    /// The actual API key value (should be hashed in production)
    /// </summary>
    public string KeyHash { get; set; } = string.Empty;

    /// <summary>
    /// Optional description of what this key is used for
    /// </summary>
    public string? Description { get; set; }

    /// <summary>
    /// Indicates whether this API key is active
    /// </summary>
    public bool IsActive { get; set; } = true;

    /// <summary>
    /// Date and time when the key was created
    /// </summary>
    public DateTime CreatedAt { get; set; }

    /// <summary>
    /// Optional expiration date for the API key
    /// </summary>
    public DateTime? ExpiresAt { get; set; }

    /// <summary>
    /// Date and time of last use
    /// </summary>
    public DateTime? LastUsedAt { get; set; }

    /// <summary>
    /// Navigation property to groups this API key can access
    /// </summary>
    public ICollection<Group> Groups { get; set; } = new List<Group>();
}
