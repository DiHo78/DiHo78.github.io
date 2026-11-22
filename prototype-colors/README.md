# PROTOTYPE: accent palettes for Stack (throwaway, never merged)

Question: which accent palette fits the blog? Five variants on the same four real posts, switchable in a bar at the
bottom, each in light and dark mode, with the measured WCAG contrast of the live colours.

## Look at it

```sh
cd prototype-colors
docker compose up
```

Then <http://localhost:1316>. Direct links: `?palette=a` … `?palette=e`, `&scheme=light` or `&scheme=dark`,
e.g. <http://localhost:1316/mein-kleiner-sqlsaturday-230-rckblick/?palette=b&scheme=dark>.
Without Docker: screenshots in [`shots/`](shots/) (taken before the last label fix: "Text sekundaer").

| | Variant | Light accent | Dark accent | Source |
|---|---|---|---|---|
| A | Blaue Stunde | `#286193` | `#67a7cc` | canal blue, the water at the coking plant |
| B | Zollverein | `#78362f` | `#db8469` | steel-girder red-brown `#5a3530`; dark: oxidised pipe `#e1866a` |
| C | Grünzug | `#3f6333` | `#8aa97a` | muted moss and olive green |
| D | Goldene Stunde | `#995b00` | `#cc9140` | amber of the small tortoiseshell `#ae701a`; dark ground: blue hour |
| E | Dompfaff | `#a94341` | `#d8847e` | bullfinch: brick-rose breast, blue-grey back (estimated, not measured) |

Anchors measured from the user's own Zollverein photos (Pillow, mean of image regions). The shared idea: a sparse,
muted ground and one living spot of colour.

## How

- `scripts/palettes.py` writes `site/assets/scss/custom.scss` and `site/data/palettes.json`. Every token is
  OKLCH; the accent's lightness moves until links, tags and hover states reach 4.5:1 on card and ground, body
  text 7:1. It exits non-zero on a failed pair. Run: `python3 -I scripts/palettes.py`.
- Roles, not swatches: ground, card, text 1–3, accent, accent hover, tag text, link underline, inline code,
  quote, table. Dark mode has its own lightness ladder, not an inversion.
- Category and tag pills: one colour (Stack hashes each name to a rainbow colour, inline, hence `!important`).
- `:focus-visible`: a 3 px ring in the accent colour.
- `layouts/_partials/head/custom.html` sets the palette before first paint; `footer/custom.html` is the bar.

## Checks

- `hugo --panicOnWarning` builds; all 100 contrast pairs (5 palettes × 2 modes × 10 pairs) pass, in Python and
  in the browser.
- `impeccable detect` 4.8.0 on the built HTML: `overused-font` (Stack's CSS names Lato, not loaded since the
  Google Fonts override), `cramped-padding` (the tweet quote's blockquote), `bounce-easing` (a Stack transition).
  No colour finding; its contrast rules need Chrome and did not run.

## Limits

- Colour only. Font, spacing and the blockquote padding are open.
- The avatar is a crop of an own photo; Stack shows it as a circle, so the watermark is not visible there.
- Five variants are a starting point: a mix (e.g. B's light with D's dark) is one line in `palettes.py`.
