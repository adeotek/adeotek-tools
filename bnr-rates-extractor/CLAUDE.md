# CLAUDE.md — bnr-rates-extractor

A .NET 10 CLI tool that fetches exchange rates from the National Bank of Romania (BNR) XML API and exports them to CSV. Single-project solution with Native AOT compilation.

## Build / Run Commands

```bash
# Build (debug)
dotnet build

# Run directly
dotnet run --project src/BnrRatesExtractor -- <year> <currency-code>

# Publish as Native AOT binary
dotnet publish src/BnrRatesExtractor -c Release -r linux-x64

# Run tests
dotnet test
```

## Usage

```bash
# Example: extract USD rates for 2024
bnr-rates-extractor 2024 USD
# Output: USD_2024_rates.csv in the current directory
```

CSV format: `Currency,Date,Rate,Multiplier` — one row per published rate day, sorted ascending, rate to 4 decimal places, `Multiplier` is 1 unless BNR quotes the currency per 100 units (HUF, JPY, KRW, ISK, IDR).

## Architecture

- `src/BnrRatesExtractor/Program.cs` — CLI entry point (top-level statements): argument validation, orchestration, exit codes.
- `src/BnrRatesExtractor/BnrRatesParser.cs` — URL construction, HTTP fetch, XML parsing, CSV rendering. Pure/static, covered by tests.
- `tests/BnrRatesExtractor.Tests/` — xUnit v3 tests (no network access; HTTP is stubbed via `HttpMessageHandler`).

Data flow: CLI args → fetch `https://curs.bnr.ro/files/xml/years/nbrfxrates{year}.xml` → parse XML → filter by currency → sort by date → write CSV.

## Key Notes / Gotchas

- **XML host**: BNR moved the XML files to `https://curs.bnr.ro` (the old `www.bnr.ro/files/xml/...` paths redirect to the homepage and return HTML). Use `BnrRatesParser.BuildYearUrl`.
- **Multiplier**: `<Rate currency="HUF" multiplier="100">` means the quoted value is RON per 100 HUF. The parser preserves it; never drop the column.
- **Native AOT**: `PublishAot=true` and `PublishTrimmed=true` are set — avoid reflection-based libraries; use source generators if serialization is added
- **InvariantGlobalization**: `InvariantGlobalization=true` is set — culture-sensitive string operations will behave differently than on a standard runtime. Formatting/parsing is explicit `CultureInfo.InvariantCulture`.
- **Year range**: BNR publishes XML for 2005 onward (`FirstSupportedYear`); later years are rejected.
- **No local caching**: each run fetches fresh XML from BNR; 30-second HTTP timeout.
- **Tests**: `dotnet test` uses the Microsoft.Testing.Platform runner (see `global.json`); .NET 10 SDK no longer supports the VSTest path.

## Code Style

- .NET 10 / C# 14 conventions
- Nullable reference types enabled (`<Nullable>enable</Nullable>`)
- Implicit usings enabled
- Top-level statements for the entry point; `readonly record struct` for data models
