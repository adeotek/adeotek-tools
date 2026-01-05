# BNR Rates Extractor

A .NET 10 CLI tool to extract exchange rates from the National Bank of Romania (BNR) and export them to CSV format.

## Features

- Fetches exchange rates from BNR's official XML API
- Extracts rates for a specific currency and year
- Outputs data in CSV format with headers: Currency, Date, Rate
- Native AOT compilation for fast startup and low memory footprint
- Modern C# 14 features with top-level statements

## Requirements

- .NET 10 SDK or later

## Building

```bash
dotnet build
```

### Building for Native AOT

To build for native AOT on Linux x64:

```bash
dotnet publish -c Release -r linux-x64
```

To build for native AOT on Windows x64:

```bash
dotnet publish -c Release -r win-x64
```

The compiled binary will be located in `bin/Release/net10.0/{runtime}/publish/`.

## Usage

```bash
bnr-rates-extractor <year> <currency-code>
```

### Arguments

- `<year>`: The year for which to extract rates (e.g., 2024)
- `<currency-code>`: The currency code to extract (e.g., USD, EUR, GBP)

### Example

```bash
bnr-rates-extractor 2024 USD
```

This will create a file named `USD_2024_rates.csv` in the current directory with the following format:

```csv
Currency,Date,Rate
USD,2024-01-02,4.4521
USD,2024-01-03,4.4532
...
```

## Output Format

The tool generates a CSV file named `{CURRENCY}_{YEAR}_rates.csv` with three columns:

- **Currency**: The currency code (e.g., USD)
- **Date**: The date in YYYY-MM-DD format
- **Rate**: The exchange rate formatted to 4 decimal places

## Error Handling

The tool validates inputs and provides meaningful error messages:

- Year must be between 2000 and the current year
- Currency code cannot be empty
- Network errors when fetching data are reported
- XML parsing errors are caught and reported

## License

See the repository LICENSE file for details.
