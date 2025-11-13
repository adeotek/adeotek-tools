namespace SslRepository.Core.Entities;

/// <summary>
/// Represents a logical grouping of certificates
/// </summary>
public class Group
{
    /// <summary>
    /// Unique identifier for the group
    /// </summary>
    public Guid Id { get; set; }

    /// <summary>
    /// Name of the group
    /// </summary>
    public string Name { get; set; } = string.Empty;

    /// <summary>
    /// Description of the group
    /// </summary>
    public string? Description { get; set; }

    /// <summary>
    /// Date and time when the group was created
    /// </summary>
    public DateTime CreatedAt { get; set; }

    /// <summary>
    /// Date and time when the group was last modified
    /// </summary>
    public DateTime UpdatedAt { get; set; }

    /// <summary>
    /// Navigation property to certificates in this group
    /// </summary>
    public ICollection<Certificate> Certificates { get; set; } = new List<Certificate>();

    /// <summary>
    /// Navigation property to API keys that have access to this group
    /// </summary>
    public ICollection<ApiKey> ApiKeys { get; set; } = new List<ApiKey>();
}
