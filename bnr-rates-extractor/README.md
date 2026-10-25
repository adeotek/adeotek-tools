# BNR Rates Extractor

A .NET 10 CLI tool to extract exchange rates from the National Bank of Romania (BNR) and export them to CSV format.

## Features

- Fetches exchange rates from BNR's official XML API (`https://curs.bnr.ro`)
- Extracts rates for a specific currency and year
- Outputs data in CSV format with headers: Currency, Date, Rate, Multiplier
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

- `<year>`: The year for which to extract rates, 2005 to the current year (e.g., 2024)
- `<currency-code>`: The 3-letter currency code to extract (e.g., USD, EUR, GBP)

### Example

```bash
bnr-rates-extractor 2024 USD
```

This will create a file named `USD_2024_rates.csv` in the current directory with the following format:

```csv
Currency,Date,Rate,Multiplier
USD,2024-01-03,4.5504,1
USD,2024-01-04,4.5395,1
...
```

## Output Format

The tool generates a CSV file named `{CURRENCY}_{YEAR}_rates.csv` with four columns:

- **Currency**: The currency code (e.g., USD)
- **Date**: The date in YYYY-MM-DD format
- **Rate**: The published rate in RON, formatted to 4 decimal places
- **Multiplier**: The number of currency units the rate applies to (1, or 100 for currencies BNR quotes per 100 units — e.g. HUF, JPY, KRW, ISK, IDR)

To get the value of one unit of the currency, divide **Rate** by **Multiplier** (e.g. 100 HUF = 1.3054 RON).

## Error Handling

The tool validates inputs and provides meaningful error messages (exit code `1`, message on stderr):

- Year must be between 2005 (the first year BNR publishes XML for) and the current year
- Currency code must be a 3-letter code (e.g. USD)
- No published rates for the requested currency/year
- Network errors (non-success HTTP status, timeout) when fetching data are reported
- Unexpected responses (HTML pages, non-BNR XML) are reported instead of silently producing an empty file

## Tests

```bash
dotnet test
```

## License

See the repository LICENSE file for details.
