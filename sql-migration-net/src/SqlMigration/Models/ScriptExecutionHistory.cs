namespace SqlMigration.Models;

public class ScriptExecutionHistory
{
    public string ScriptFile { get; set; } = null!;
    public string ScriptHash { get; set; } = null!;
    public DateTime ExecutedAt { get; set; } = DateTime.UtcNow;
}
