# PROTOTYPE: accent palettes for Stack (throwaway, never merged)

Question: which accent palette fits the blog? Eight variants on the same four real posts, switchable in a bar at the
bottom, each in light and dark mode, with the measured WCAG contrast of the live colours.

## Look at it

```sh
cd prototype-colors
docker compose up
```

Then <http://localhost:1316>. Direct links: `?palette=a` … `?palette=h`, `&scheme=light` or `&scheme=dark`,
e.g. <http://localhost:1316/mein-kleiner-sqlsaturday-230-rckblick/?palette=b&scheme=dark>.
Without Docker: screenshots in [`shots/`](shots/).

| | Variant | Light accent | Dark accent | Source |
|---|---|---|---|---|
| A | Blaue Stunde | `#01519f` | `#4fa8e1` | deep canal blue (second round: more saturated, night-blue dark ground) |
| B | Zollverein | `#78362f` | `#db8469` | steel-girder red-brown `#5a3530`; dark: oxidised pipe `#e1866a` |
| C | Grünzug | `#3f6333` | `#8aa97a` | muted moss and olive green |
| D | Goldene Stunde | `#995b00` | `#cc9140` | amber of the small tortoiseshell `#ae701a`; dark ground: blue hour |
| E | Dompfaff | `#a94341` | `#d8847e` | bullfinch: brick-rose breast, blue-grey back (estimated, not measured) |
| F | Schiefer | `#3f4f60` | `#8fa1b0` | slate and steel grey as the only accent |
| G | Schiefer & Glut | `#3a495a` + amber glow `#f8ac3d` | `#8fa1b0` + amber glow `#f3af49` | F's slate frame; amber as a bright surface: tag pills (coal-dark text), link underline, edge of headings and quotes |
| H | Kohle & Stahl | `#364451` + ember glow `#f55e47` on a coal frame | `#91a1ad` + ember glow `#f56249` | "Auf Kohle geboren, im Herzen aus Stahl": coal-black ground and sidebar also in light mode, light steel cards, red ember on pills, links, site name, edges |

Anchors measured from the user's own Zollverein photos (Pillow, mean of image regions). The shared idea: a sparse,
muted ground and one living spot of colour. G and H carry it literally: a glow (`glow`) as a bright surface, which may stay bright because its text is coal-dark,
not as thin link text, which on white must be dark enough to look brown (round 2 lesson: the user saw no glow there).

## Decision

The user picked **H Kohle & Stahl** as the starting palette, with the ember shifted towards red (round 4).

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

- `hugo --panicOnWarning` builds; all 208 contrast pairs (8 palettes × 2 modes × 13 pairs) pass, in Python and
  in the browser.
- `impeccable detect` 4.8.0 on the built HTML: `overused-font` (Stack's CSS names Lato, not loaded since the
  Google Fonts override), `cramped-padding` (the tweet quote's blockquote), `bounce-easing` (a Stack transition).
  No colour finding; its contrast rules need Chrome and did not run.

## Limits

- Colour only. Font, spacing and the blockquote padding are open.
- The avatar is a crop of an own photo; Stack shows it as a circle, so the watermark is not visible there.
- Eight variants are a starting point: a mix (e.g. B's light with D's dark) is one line in `palettes.py`.
