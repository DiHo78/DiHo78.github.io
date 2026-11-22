# PROTOTYPE, throwaway.
# Keep channel meta, taxonomies, the chosen posts and their attachments; drop all comments and author e-mails.
import sys, urllib.parse
import xml.etree.ElementTree as ET
src, dst = sys.argv[1], sys.argv[2]; prefixes = sys.argv[3:]
NS = {'wp':'http://wordpress.org/export/1.2/','content':'http://purl.org/rss/1.0/modules/content/',
      'excerpt':'http://wordpress.org/export/1.2/excerpt/','dc':'http://purl.org/dc/elements/1.1/',
      'wfw':'http://wellformedweb.org/CommentAPI/','atom':'http://www.w3.org/2005/Atom'}
for k,v in NS.items(): ET.register_namespace(k,v)
W='{%s}'%NS['wp']
tree = ET.parse(src); ch = tree.getroot().find('channel')
items = ch.findall('item'); keep=set()
for it in items:
    slug = urllib.parse.unquote(it.findtext(W+'post_name') or '')
    if it.findtext(W+'post_type')=='post' and it.findtext(W+'status')=='publish' and any(slug.startswith(p) for p in prefixes):
        keep.add(it.findtext(W+'post_id'))
for it in items:
    t = it.findtext(W+'post_type'); pid=it.findtext(W+'post_id'); par=it.findtext(W+'post_parent')
    if not ((t=='post' and pid in keep) or (t=='attachment' and par in keep)):
        ch.remove(it); continue
    for c in it.findall(W+'comment'): it.remove(c)
for a in ch.findall(W+'author'):
    e=a.find(W+'author_email')
    if e is not None: e.text=''
tree.write(dst, encoding='utf-8', xml_declaration=True)
print('kept posts', len(keep), 'items', len(ch.findall('item')))
