namespace SqlMigration.Models;

public class ConnectionOptions
{
    public string? Provider { get; set; }
    public string? ConnectionString { get; set; }
    public string? Host { get; set; }
    public int? Port { get; set; }
    public string? Database { get; set; }
    public string? User { get; set; }
    public string? Password { get; set; }
}
