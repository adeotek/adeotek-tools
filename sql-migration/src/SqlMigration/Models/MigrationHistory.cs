namespace SqlMigration.Models;

public class MigrationHistory
{
    public int Id { get; set; }
    public string ScriptName { get; set; } = string.Empty;
    public string Hash { get; set; } = string.Empty;
    public DateTime ExecutedAt { get; set; }
}
