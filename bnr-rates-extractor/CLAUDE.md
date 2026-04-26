# CLAUDE.md — bnr-rates-extractor

A .NET 10 CLI tool that fetches exchange rates from the National Bank of Romania (BNR) XML API and exports them to CSV. Single-project solution with Native AOT compilation.

## Build / Run Commands

```bash
# Build (debug)
dotnet build

# Run directly
dotnet run --project src/BnrRatesExtractor -- <year> <currency-code>

# Publish as Native AOT binary
dotnet publish src/BnrRatesExtractor -c Release

# Run tests (none currently defined)
dotnet test
```

## Usage

```bash
# Example: extract USD rates for 2024
bnr-rates-extractor 2024 USD
# Output: USD_2024_rates.csv in the current directory
```

CSV format: `Currency,Date,Rate` — one row per published rate day, sorted ascending, rate to 4 decimal places.

## Architecture

Single-file implementation (`src/BnrRatesExtractor/Program.cs`) using C# top-level statements. No external dependencies beyond the .NET runtime.

Data flow: CLI args → fetch `https://www.bnr.ro/files/xml/years/nbrfxrates{year}.xml` → parse XML → filter by currency → sort by date → write CSV.

## Key Notes / Gotchas

- **Native AOT**: `PublishAot=true` and `PublishTrimmed=true` are set — avoid reflection-based libraries; use source generators if serialization is added
- **InvariantGlobalization**: `InvariantGlobalization=true` is set — culture-sensitive string operations will behave differently than on a standard runtime
- **Year range**: hardcoded validation accepts 2000 to current year; BNR XML URLs follow the pattern `nbrfxrates{year}.xml`
- **No local caching**: each run fetches fresh XML from BNR; 30-second HTTP timeout

## Code Style

- .NET 10 / C# 14 conventions
- Nullable reference types enabled (`<Nullable>enable</Nullable>`)
- Implicit usings enabled
- Top-level statements for the entry point; `readonly record struct` for data models
