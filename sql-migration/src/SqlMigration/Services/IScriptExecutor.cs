namespace SqlMigration.Services;

public interface IScriptExecutor
{
    Task ExecuteScript(string connectionString, string scriptContent);
}
