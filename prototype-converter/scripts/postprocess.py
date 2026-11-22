# PROTOTYPE, throwaway: turn wp2hugo output into the target shape of ADR 0001.
# usage: python3 -I postprocess.py <wp2hugo site dir> <filtered export.xml> <out content dir>
import html, json, os, re, shutil, sys, urllib.parse, urllib.request
import xml.etree.ElementTree as ET

src, xml_path, out = sys.argv[1:4]
UA = {'User-Agent': 'agentos-consultant'}
WP = 'https://dirkhondong.wordpress.com'
W = '{http://wordpress.org/export/1.2/}'

# taxonomy slug -> display name, from the export (decision 14: keep the WordPress slugs in the paths)
names = {}
for it in ET.parse(xml_path).getroot().iter('item'):
    for c in it.findall('category'):
        names[(c.get('domain'), c.get('nicename'))] = c.text

def fetch(url):
    with urllib.request.urlopen(urllib.request.Request(url, headers=UA), timeout=30) as r:
        return r.read()

def front_matter(text):
    head, body = text.split('\n---\n', 1)
    fm, key = {}, None
    for line in head.splitlines()[1:]:
        m = re.match(r'^([A-Za-z_][\w]*):\s*(.*)$', line)
        if m:
            key, val = m.group(1), m.group(2)
            fm[key] = [] if val == '' else json.loads(val) if val.startswith('"') else val.strip("'")
        elif line.startswith('  - ') and isinstance(fm.get(key), list):
            fm[key].append(line[4:].strip('"'))
    return fm, body

def guess_lang(body):
    words = re.findall(r'[a-zäöüß]+', body.lower())
    de = sum(w in {'und', 'der', 'die', 'das', 'ich', 'nicht', 'ist', 'mit', 'auch'} for w in words)
    en = sum(w in {'the', 'and', 'is', 'to', 'of', 'you', 'that', 'with', 'it'} for w in words)
    return 'de' if de > en else 'en'

def code_lang(code):
    if re.search(r'\b(SELECT|CREATE|DROP|ALTER|EXEC|PROCEDURE)\b', code):
        return 'sql'
    if re.search(r'\$\w+|\b[A-Z][a-z]+-[A-Z]\w+', code):
        return 'powershell'
    return 'text'

def merge_live_writer(body):
    # Live Writer wrote one <pre> per code line; wp2hugo makes each one a fence.
    # Merge fences that only blank lines separate, and drop the "  N: " line numbers.
    items, lines, i = [], body.split('\n'), 0
    while i < len(lines):
        if lines[i].startswith('```'):
            j = i + 1
            while j < len(lines) and lines[j] != '```':
                j += 1
            items.append(('code', lines[i + 1:j]))
            i = j + 1
        else:
            items.append(('text', lines[i]))
            i += 1
    out, k = [], 0
    while k < len(items):
        if items[k][0] != 'code':
            out.append(items[k][1]); k += 1; continue
        code, k = list(items[k][1]), k + 1
        while True:
            n = k
            while n < len(items) and items[n] == ('text', ''):
                n += 1
            if n < len(items) and items[n][0] == 'code':
                code += items[n][1]; k = n + 1
            else:
                break
        if code and re.match(r'^\s*\d+:', code[0]):
            code = [re.sub(r'^\s*\d+: ?', '', l) for l in code]
        text = '\n'.join(code)
        out += ['```' + code_lang(text)] + code + ['```']
    return '\n'.join(out)

def gist(m):
    # decision 12: a Gist becomes a fenced code block per file plus a link to the Gist
    gid = m.group(2)
    data = json.loads(fetch('https://api.github.com/gists/' + gid))
    parts = []
    for name, f in data['files'].items():
        lang = (f.get('language') or 'text').lower()
        parts.append('```%s\n%s\n```' % (lang, f['content'].rstrip('\n')))
    parts.append('[Gist %s](https://gist.github.com/%s/%s)' % (gid[:7], m.group(1), gid))
    return '\n\n'.join(parts)

def bundle_media(body, bundle):
    # decision 15: every image lives next to the post
    def local(path):
        name = os.path.basename(path)
        dst = os.path.join(bundle, name)
        if not os.path.exists(dst):
            have = os.path.join(src, 'static', path.lstrip('/'))
            if os.path.exists(have):
                shutil.copy(have, dst)
            else:  # wp2hugo fetched only the <img> src, not the full-size link target
                open(dst, 'wb').write(fetch(WP + path))
        return name
    body = re.sub(r'\((?:https?://dirkhondong\.(?:wordpress|files\.wordpress)\.com)?(/wp-content/uploads/[^)\s?"]+)(?:\?[^)\s"]*)?',
                  lambda m: '(' + local(m.group(1)), body)
    # decision 13: own Flickr photos get a local preview; farmN.staticflickr.com is served as live.staticflickr.com
    def flickr(m):
        name = m.group(2)
        dst = os.path.join(bundle, name)
        if not os.path.exists(dst):
            open(dst, 'wb').write(fetch('https://live.staticflickr.com/%s/%s' % (m.group(1), name)))
        return '(' + name
    return re.sub(r'\(https?://farm\d+\.staticflickr\.com/(\d+)/([^)\s"]+)', flickr, body)

os.makedirs(out, exist_ok=True)
for fn in sorted(os.listdir(os.path.join(src, 'content', 'posts'))):
    fm, body = front_matter(open(os.path.join(src, 'content', 'posts', fn), encoding='utf-8').read())
    slug = urllib.parse.unquote(fm['url']).strip('/').split('/')[-1]
    bundle = os.path.join(out, 'posts', slug)
    os.makedirs(bundle, exist_ok=True)
    body = merge_live_writer(body)
    body = re.sub(r'^https://gist\.github\.com/([\w-]+)/([0-9a-f]+)[ \t]*$', gist, body, flags=re.M)
    body = bundle_media(body, bundle)
    # decision 7: internal links /YYYY/MM/DD/<slug>/ -> /<slug>/
    body = re.sub(r'\]\((?:https?://dirkhondong\.wordpress\.com)?/\d{4}/\d{2}/\d{2}/([^/)]+)/?\)', r'](/\1/)', body)
    meta = {
        'title': html.unescape(fm['title']),
        'date': fm['date'],
        # decision 7 is /:slug/, but Hugo's slug urlizes and drops "…" and "´"; url keeps the WordPress slug as is
        'url': '/%s/' % slug,
        'lang': guess_lang(body),
        'categories': [names.get(('category', c), c) for c in fm.get('categories', []) if c != 'uncategorized'],
        'tags': [names.get(('post_tag', t), t) for t in fm.get('tags', [])],
    }
    lines = ['---'] + ['%s: %s' % (k, json.dumps(v, ensure_ascii=False)) for k, v in meta.items()] + ['---', '']
    open(os.path.join(bundle, 'index.md'), 'w', encoding='utf-8').write('\n'.join(lines) + body.lstrip('\n'))
    print('wrote', bundle)
