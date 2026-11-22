# PROTOTYPE, throwaway: derive the palettes and write site/assets/scss/custom.scss + site/data/palettes.json.
# usage (from prototype-colors/): python3 -I scripts/palettes.py
# Every token is OKLCH; lightness is nudged until each contrast pair in CHECKS meets its WCAG target.
import json, os, sys
sys.path.insert(0, os.path.dirname(__file__))
from oklch import oklch, contrast

# hue/chroma come from the measured photo anchors (REPORT.md); the bullfinch is estimated, not measured
VARIANTS = [
    dict(id='a', light_L=0.44, name='Blaue Stunde', origin='Tiefes Kanalblau, satter als zuvor; dunkel: Nachtblau mit Wasserblau',
         neutral=(250, 0.016), accent=(255, 0.150), dark_neutral=(258, 0.040), dark_accent=(240, 0.120)),
    dict(id='b', light_L=0.42, name='Zollverein', origin='Stahlträger-Rotbraun (#5a3530); dunkel: oxidiertes Rohr (#e1866a)',
         neutral=(60, 0.008), accent=(28, 0.095), dark_neutral=(60, 0.010), dark_accent=(38, 0.115)),
    dict(id='c', light_L=0.46, name='Grünzug', origin='Gedecktes Moos- und Olivgrün; Felder, Wiesen, Wälder vom Rad aus',
         neutral=(135, 0.009), accent=(138, 0.085), dark_neutral=(155, 0.016), dark_accent=(135, 0.075)),
    dict(id='d', light_L=0.53, name='Goldene Stunde', origin='Bernstein vom Kleinen Fuchs (#ae701a) auf kühlem Grau; dunkel: blaue Stunde',
         neutral=(250, 0.006), accent=(66, 0.120), dark_neutral=(255, 0.026), dark_accent=(72, 0.120)),
    dict(id='e', light_L=0.52, name='Dompfaff', origin='Brust ziegelrosa, Rücken blaugrau, Kappe schwarz (geschätzt)',
         neutral=(250, 0.014), accent=(24, 0.135), dark_neutral=(250, 0.018), dark_accent=(24, 0.105)),
    dict(id='f', light_L=0.42, name='Schiefer', origin='Schiefer und Stahlgrau als einziger Akzent; kühl, schwer, ohne Farbaussage',
         neutral=(250, 0.006), accent=(252, 0.035), dark_neutral=(250, 0.010), dark_accent=(245, 0.030)),
    dict(id='g', light_L=0.40, name='Schiefer & Glut', origin='Schiefer wie F, dazu Bernstein als leuchtende Fläche: Tags, Unterstreichung, Kante an Überschriften und Zitaten',
         neutral=(250, 0.006), accent=(252, 0.035), dark_neutral=(252, 0.012), dark_accent=(245, 0.030),
         glow=(72, 0.150, 0.80), dark_glow=(74, 0.140, 0.80), dark_spot=(74, 0.130, 0.80)),
    dict(id='h', light_L=0.38, name='Kohle & Stahl', origin='„Auf Kohle geboren, im Herzen aus Stahl“: Kohle als Rahmen auch im hellen Modus, Stahlblech-Karten, rote Glut auf Tags, Links und Kanten',
         neutral=(245, 0.008), accent=(245, 0.028), dark_neutral=(250, 0.006), dark_accent=(240, 0.025),
         glow=(31, 0.190, 0.66), dark_glow=(32, 0.185, 0.68), spot=(29, 0.175, 0.52), dark_spot=(32, 0.170, 0.72),
         coal=True, dark_ground=0.155),
]

# (foreground, background, minimum ratio, label)
CHECKS = [
    ('card-text-color-main', 'card-background', 7.0, 'Text auf Karte'),
    ('card-text-color-secondary', 'card-background', 4.5, 'Text sekundär'),
    ('card-text-color-tertiary', 'card-background', 4.5, 'Metadaten'),
    ('body-text-color', 'body-background', 4.5, 'Seitenleiste'),
    ('proto-link', 'card-background', 4.5, 'Link auf Karte'),
    ('proto-ground-accent', 'body-background', 4.5, 'Akzent auf Grund'),
    ('proto-link-hover', 'card-background', 4.5, 'Link Hover'),
    ('accent-color', 'card-background', 4.5, 'Akzent auf Karte'),
    ('accent-color-text', 'accent-color', 4.5, 'Knopf-Text'),
    ('accent-color-text', 'accent-color-darker', 4.5, 'Knopf Hover'),
    ('proto-glow-text', 'proto-glow', 4.5, 'Tag-Text'),
    ('proto-glow-text', 'proto-glow-hover', 4.5, 'Tag Hover'),
    ('code-text-color', 'code-background-color', 4.5, 'Inline-Code'),
]

def on_ground(t, v, dark_mode):
    # a colour used on the page ground must pass there too, except on H's coal ground in light mode
    return dark_mode or not v.get('coal')

def spot(t, v, key, dark_mode):
    # optional second colour for link text; otherwise links use the accent
    if key not in v:
        t['proto-link'], t['proto-link-hover'] = t['accent-color'], t['accent-color-darker']
        return
    h, c, L = v[key]
    while True:
        t['proto-link'] = oklch(L, c, h)
        t['proto-link-hover'] = oklch(L + (0.06 if dark_mode else -0.07), c, h)
        if (contrast(t['proto-link'], t['card-background']) >= 4.6 and
                (not on_ground(t, v, dark_mode) or contrast(t['proto-link'], t['body-background']) >= 4.6)):
            break
        L += 0.005 if dark_mode else -0.005

def glow(t, v, key, dark_mode, nh):
    # optional glowing fill (palettes G, H): tag pills, link underline, heading and quote edge.
    # It is a surface, so it may stay bright; the pill text is coal-dark and must reach 4.5:1 on it.
    if key not in v:
        t['proto-glow'], t['proto-glow-hover'] = t['accent-color'], t['accent-color-darker']
        t['proto-glow-text'] = t['accent-color-text']
        t['proto-ground-accent'] = t['accent-color'] if on_ground(t, v, dark_mode) else t['card-text-color-main']
        return None
    h, c, L = v[key]
    t['proto-glow-text'] = oklch(0.20, 0.012, nh)
    while True:
        t['proto-glow'] = oklch(L, c, h)
        t['proto-glow-hover'] = oklch(L - 0.05, c, h)
        if contrast(t['proto-glow-text'], t['proto-glow-hover']) >= 4.6:
            break
        L += 0.005
    # site name, current menu item, footer: on the ground, the glow where it reads, else the accent
    t['proto-ground-accent'] = (t['proto-glow'] if contrast(t['proto-glow'], t['body-background']) >= 4.6
                                else t['accent-color'])
    return (h, c)

def light(v):
    nh, nc = v['neutral']; ah, ac = v['accent']
    t = {
        'body-background': oklch(0.945, nc, nh),
        'card-background': oklch(0.995, nc * 0.3, nh),
        'card-background-selected': oklch(0.925, nc * 1.5, nh),
        'card-text-color-main': oklch(0.24, nc * 1.5, nh),
        'card-text-color-secondary': oklch(0.42, nc * 1.5, nh),
        'card-text-color-tertiary': oklch(0.48, nc * 1.5, nh),
        'card-separator-color': oklch(0.90, nc, nh),
        'body-text-color': oklch(0.40, nc * 1.5, nh),
        'accent-color-text': '#ffffff',
        'code-background-color': oklch(0.93, ac * 0.18, ah),
        'code-text-color': oklch(0.36, ac * 0.6, ah),
        'blockquote-background-color': oklch(0.965, nc, nh),
        'tr-even-background-color': oklch(0.965, nc, nh),
        'table-border-color': oklch(0.86, nc, nh),
    }
    if v.get('coal'):  # coal frame: dark ground and sidebar around light steel cards
        t['body-background'] = oklch(0.215, nc, nh)
        t['body-text-color'] = oklch(0.84, nc, nh)
        t['card-background'] = oklch(0.985, nc * 0.6, nh)
    L = v['light_L']
    while True:  # darken the accent until links and tags pass on the card (and on the ground)
        t['accent-color'] = oklch(L, ac, ah)
        t['accent-color-darker'] = oklch(L - 0.07, ac, ah)
        if (contrast(t['accent-color'], t['card-background']) >= 4.6 and
                (not on_ground(t, v, False) or contrast(t['accent-color'], t['body-background']) >= 4.6) and
                contrast(t['accent-color-text'], t['accent-color']) >= 4.6):
            break
        L -= 0.005
    spot(t, v, 'spot', False)
    gl = glow(t, v, 'glow', False, nh)
    uh, uc = gl or (ah, ac)
    t['link-underline'] = oklch(0.84, uc * 0.45, uh)
    return t, L

def dark(v):
    nh, nc = v['dark_neutral']; ah, ac = v['dark_accent']
    t = {
        'body-background': oklch(v.get('dark_ground', 0.195), nc, nh),
        'card-background': oklch(v.get('dark_ground', 0.195) + 0.05, nc, nh),
        'card-background-selected': oklch(0.30, nc, nh),
        'card-text-color-main': oklch(0.94, nc * 0.4, nh),
        'card-text-color-secondary': oklch(0.82, nc * 0.5, nh),
        'card-text-color-tertiary': oklch(0.74, nc * 0.5, nh),
        'card-separator-color': oklch(0.33, nc, nh),
        'body-text-color': oklch(0.80, nc * 0.5, nh),
        'accent-color-text': oklch(0.18, nc, nh),
        'code-background-color': oklch(0.17, nc, nh),
        'code-text-color': oklch(0.88, ac * 0.4, ah),
        'blockquote-background-color': oklch(0.28, nc, nh),
        'tr-even-background-color': oklch(0.28, nc, nh),
        'table-border-color': oklch(0.40, nc, nh),
    }
    L = 0.70
    while True:  # lighten the accent until links pass on the card and its hover still carries dark tag text
        t['accent-color'] = oklch(L, ac, ah)
        t['accent-color-darker'] = oklch(L - 0.06, ac, ah)
        if (contrast(t['accent-color'], t['card-background']) >= 4.6 and
                contrast(t['accent-color-text'], t['accent-color-darker']) >= 4.6):
            break
        L += 0.005
    spot(t, v, 'dark_spot', True)
    gl = glow(t, v, 'dark_glow', True, nh)
    uh, uc = gl or (ah, ac)
    t['link-underline'] = oklch(0.45, uc * 0.6, uh)
    return t, L

def rgb_triple(h):
    h = h.lstrip('#'); return ', '.join(str(int(h[i:i + 2], 16)) for i in (0, 2, 4))

def block(sel, t):
    lines = [sel + ' {']
    for k, v in t.items():
        if k == 'link-underline':
            lines.append(f'    --link-background-color: {rgb_triple(v)};')
        else:
            lines.append(f'    --{k}: {v};')
    return '\n'.join(lines + ['}'])

out, data, fail = [], [], 0
for v in VARIANTS:
    for scheme, (t, L) in (('light', light(v)), ('dark', dark(v))):
        sel = f':root[data-palette="{v["id"]}"]' + ('[data-scheme="dark"]' if scheme == 'dark' else '')
        out.append(block(sel, t))
        for fg, bg, target, label in CHECKS:
            r = contrast(t[fg], t[bg])
            if r < target:
                fail += 1
                print(f'FAIL {v["id"]} {scheme} {label}: {r:.2f} < {target}')
        print(f'{v["id"]} {scheme:5s} accent {t["accent-color"]} (L={L:.3f}) on card {contrast(t["accent-color"], t["card-background"]):.2f}')
    data.append({k: v[k] for k in ('id', 'name', 'origin')})

COMMON = '''
/* Single-colour category and tag pills instead of Stack's hashed rainbow (inline style, hence !important) */
.article-category a, .article-tags a {
    background-color: var(--proto-glow) !important;
    color: var(--proto-glow-text) !important;
    &:hover { background-color: var(--proto-glow-hover) !important; }
}

/* Accent on the page ground: on H's coal frame it must be light (proto-ground-accent) */
.sidebar .site-name, .sidebar .site-name a, #main-menu li.current a, footer.site-footer .copyright { color: var(--proto-ground-accent); }

/* Link text and focus take the optional second colour (G dark, H); elsewhere it equals the accent */
.article-content a, .widget a.link { color: var(--proto-link); }
.article-content a:hover { color: var(--proto-link-hover); }

/* Visible keyboard focus in every palette (WCAG 2.4.7) */
:focus-visible {
    outline: 3px solid var(--proto-link);
    outline-offset: 2px;
    border-radius: 2px;
}
.sidebar :focus-visible, footer.site-footer :focus-visible { outline-color: var(--proto-ground-accent); }

/* Room for the prototype's palette bar */
body { padding-bottom: 140px; }
'''

GLOW = [v['id'] for v in VARIANTS if 'glow' in v]
sel = ', '.join(f':root[data-palette="{i}"]' for i in GLOW)
COMMON += f'''
/* Palettes with a glow ({', '.join(GLOW).upper()}): a solid, bright underline and a glowing edge on headings and quotes */
:is({sel}) .article-content a {{
    box-shadow: inset 0 -2px 0 var(--proto-glow);
    &:hover {{ box-shadow: inset 0 -4px 0 var(--proto-glow); }}
}}
:is({sel}) .article-content :is(h1, h2, h3, h4, h5, h6),
:is({sel}) .article-content blockquote {{ border-inline-start-color: var(--proto-glow) !important; }}
'''

head = '// PROTOTYPE, generated by scripts/palettes.py: do not edit by hand.\n'
os.makedirs('site/assets/scss', exist_ok=True); os.makedirs('site/data', exist_ok=True)
open('site/assets/scss/custom.scss', 'w', encoding='utf-8').write(head + '\n\n'.join(out) + '\n' + COMMON)
json.dump({'variants': data, 'checks': [dict(fg=f, bg=b, target=t, label=l) for f, b, t, l in CHECKS]},
          open('site/data/palettes.json', 'w', encoding='utf-8'), ensure_ascii=False, indent=1)
print('contrast failures:', fail)
sys.exit(1 if fail else 0)
