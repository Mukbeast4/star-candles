# star-candles

GitHub star history rendered as a Japanese candlestick chart, updated hourly by a GitHub Action.

<details open>
  <summary><b>Daily</b></summary>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Mukbeast4/star-candles/chart/chart-daily-dark.svg">
    <img alt="Daily candlestick chart of this repository's GitHub stars" src="https://raw.githubusercontent.com/Mukbeast4/star-candles/chart/chart-daily-light.svg">
  </picture>
</details>

<details>
  <summary><b>Monthly</b></summary>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Mukbeast4/star-candles/chart/chart-monthly-dark.svg">
    <img alt="Monthly candlestick chart of this repository's GitHub stars" src="https://raw.githubusercontent.com/Mukbeast4/star-candles/chart/chart-monthly-light.svg">
  </picture>
</details>

<details>
  <summary><b>Yearly</b></summary>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Mukbeast4/star-candles/chart/chart-yearly-dark.svg">
    <img alt="Yearly candlestick chart of this repository's GitHub stars" src="https://raw.githubusercontent.com/Mukbeast4/star-candles/chart/chart-yearly-light.svg">
  </picture>
</details>

## Add it to your repository

Create `.github/workflows/chart.yml` in your repository:

```yaml
name: chart

on:
  schedule:
    - cron: "23 * * * *"
  workflow_dispatch:

permissions:
  contents: write

concurrency:
  group: chart

jobs:
  chart:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: Mukbeast4/star-candles@main
```

Then embed the charts in your README, replacing `OWNER/REPO`:

```html
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/OWNER/REPO/chart/chart-daily-dark.svg">
  <img alt="Daily candlestick chart of this repository's GitHub stars" src="https://raw.githubusercontent.com/OWNER/REPO/chart/chart-daily-light.svg">
</picture>
```

The same URLs exist for `chart-monthly-*.svg` and `chart-yearly-*.svg`; wrap each in a `<details>` block to get the switcher shown above. The first run creates the `chart` branch on its own, backfills the star history, and pushes the six SVGs; run the workflow once by hand from the Actions tab if you do not want to wait for the cron. Inputs available on the action: `chart-branch` (default `chart`), `max-candles` (default `90`), and `token` (defaults to the workflow token; used for both API calls and the chart branch push, so `persist-credentials: false` on the checkout is fine, and passing a PAT makes chart pushes able to trigger other workflows).

## How it works

An hourly workflow samples the repository's current star count, appends it to a JSON history, aggregates the samples into OHLC candles at three granularities (one candle per day, per month, or per year), and renders an SVG chart per granularity in light and dark. The generated files live on the `chart` branch, created automatically on the first run, so the default branch only ever contains code; the README embeds them through `raw.githubusercontent.com` and the blocks above switch between the daily, monthly and yearly views. If the chart branch is ever deleted, the next run recreates it and backfills again.

On the very first run the history is backfilled from the GitHub stargazers API, which exposes the timestamp of every star. GitHub only serves that listing for repositories the token has access to, so backfill works for the repository's own workflow token but not for arbitrary third-party repositories; when it is unavailable the history simply starts from the first sample. Backfilled days only ever go up because unstars are invisible in that API; once hourly sampling takes over, days where unstars outnumber stars show up as red candles.

## Reading the chart

- Each candle covers one day, one month or one year depending on the selected view.
- Hollow green candle: stars closed the period higher than they opened.
- Filled red candle: stars closed the period lower (more unstars than stars).
- Wicks mark the period's high and low, captured by the hourly samples.
- The candle direction is encoded twice, by color and by hollow versus filled body, so the chart stays readable for colorblind readers and in grayscale.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `REPO` | `$GITHUB_REPOSITORY` | Repository to track, as `owner/name` |
| `GITHUB_TOKEN` | none | Token for API calls, optional but avoids rate limits |
| `GITHUB_API_URL` | `https://api.github.com` | API base URL |
| `HISTORY_FILE` | `out/history.json` | Path of the sample history |
| `OUTPUT_DIR` | `out` | Directory for the six generated SVGs |
| `MAX_CANDLES` | `90` | Number of most recent candles rendered per view |

## Running locally

```bash
REPO=Mukbeast4/star-candles GITHUB_TOKEN=$(gh auth token) go run ./cmd/star-candles
```

The charts land in `out/chart-{daily,monthly,yearly}-{light,dark}.svg`. Point `REPO` at any public repository to render its history.

## Notes

- The scheduled workflow only runs once this code is on the default branch.
- GitHub suspends scheduled workflows after 60 days without repository activity; the `workflow_dispatch` trigger revives the chart manually if that happens.
- The stargazers API returns at most 40,000 stars, so backfill history is capped there.
