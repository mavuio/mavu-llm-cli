---
name: beans-stats
description: Generate stats and summaries from beans metadata
---

# Beans Stats

Use this skill to compute summary stats from `.beans/*.md` front matter (status, tags, parents, titles).

## Workflow

1. Identify scope (tag/type/status) with `beans list` or `beans list -S`.
2. Parse `.beans` front matter with a small Python script (no external deps).
3. Compute requested stats (status counts, severity rollups, child counts).
4. Return concise results.

## Base parser

```bash
python - <<'PY'
import glob
import re
from collections import defaultdict, Counter

def parse_front_matter(path):
    lines = open(path, 'r', encoding='utf-8').read().splitlines()
    if not lines or lines[0].strip() != '---':
        return None
    data = {}
    tags = []
    i = 1
    while i < len(lines):
        line = lines[i]
        if line.strip() == '---':
            break
        if line.strip() == 'tags:':
            i += 1
            while i < len(lines) and re.match(r'^\s*-\s+', lines[i]):
                tags.append(re.sub(r'^\s*-\s+', '', lines[i]).strip())
                i += 1
            data['tags'] = tags
            continue
        m = re.match(r'^(\w+):\s*(.*)$', line)
        if m:
            data[m.group(1)] = m.group(2).strip().strip('"').strip("'")
        i += 1
    return data

beans = []
for path in glob.glob('.beans/*.md'):
    data = parse_front_matter(path)
    if not data:
        continue
    data['id'] = path.split('/')[-1].split('--')[0]
    beans.append(data)

# Customize below for the requested stats.
PY
```

## Common stats recipes

### Unique tags

```bash
python - <<'PY'
import glob
import re

def parse_front_matter(path):
    lines = open(path, 'r', encoding='utf-8').read().splitlines()
    if not lines or lines[0].strip() != '---':
        return None
    data = {}
    tags = []
    i = 1
    while i < len(lines):
        line = lines[i]
        if line.strip() == '---':
            break
        if line.strip() == 'tags:':
            i += 1
            while i < len(lines) and re.match(r'^\s*-\s+', lines[i]):
                tags.append(re.sub(r'^\s*-\s+', '', lines[i]).strip())
                i += 1
            data['tags'] = tags
            continue
        i += 1
    return data

tags = set()
for path in glob.glob('.beans/*.md'):
    data = parse_front_matter(path)
    if data:
        for tag in data.get('tags', []):
            tags.add(tag)

print('\n'.join(sorted(tags)))
PY
```

### Status by severity (from warning titles)

```bash
python - <<'PY'
import glob
import re
from collections import defaultdict, Counter

def parse_front_matter(path):
    lines = open(path, 'r', encoding='utf-8').read().splitlines()
    if not lines or lines[0].strip() != '---':
        return None
    data = {}
    tags = []
    i = 1
    while i < len(lines):
        line = lines[i]
        if line.strip() == '---':
            break
        if line.strip() == 'tags:':
            i += 1
            while i < len(lines) and re.match(r'^\s*-\s+', lines[i]):
                tags.append(re.sub(r'^\s*-\s+', '', lines[i]).strip())
                i += 1
            data['tags'] = tags
            continue
        m = re.match(r'^(\w+):\s*(.*)$', line)
        if m:
            data[m.group(1)] = m.group(2).strip().strip('"').strip("'")
        i += 1
    return data

beans = []
for path in glob.glob('.beans/*.md'):
    data = parse_front_matter(path)
    if data:
        beans.append(data)

severity_status = defaultdict(Counter)
for b in beans:
    if b.get('type') != 'epic':
        continue
    title = b.get('title', '')
    m = re.search(r'Warnings:\s*([^|]+)\|', title)
    severity = m.group(1).strip() if m else 'Unknown'
    severity_status[severity][b.get('status', 'unknown')] += 1

for severity in sorted(severity_status):
    parts = ', '.join(f"{k}:{v}" for k, v in sorted(severity_status[severity].items()))
    print(f"{severity}\t{parts}")
PY
```

### Child counts by parent epic

```bash
python - <<'PY'
import glob
import re
from collections import defaultdict, Counter

def parse_front_matter(path):
    lines = open(path, 'r', encoding='utf-8').read().splitlines()
    if not lines or lines[0].strip() != '---':
        return None
    data = {}
    tags = []
    i = 1
    while i < len(lines):
        line = lines[i]
        if line.strip() == '---':
            break
        if line.strip() == 'tags:':
            i += 1
            while i < len(lines) and re.match(r'^\s*-\s+', lines[i]):
                tags.append(re.sub(r'^\s*-\s+', '', lines[i]).strip())
                i += 1
            data['tags'] = tags
            continue
        m = re.match(r'^(\w+):\s*(.*)$', line)
        if m:
            data[m.group(1)] = m.group(2).strip().strip('"').strip("'")
        i += 1
    return data

beans = {}
for path in glob.glob('.beans/*.md'):
    data = parse_front_matter(path)
    if not data:
        continue
    bean_id = path.split('/')[-1].split('--')[0]
    data['id'] = bean_id
    beans[bean_id] = data

children_by_parent = defaultdict(list)
for b in beans.values():
    parent = b.get('parent')
    if parent:
        children_by_parent[parent].append(b)

for b in sorted(beans.values(), key=lambda x: x['id']):
    if b.get('type') == 'epic' and not b.get('parent'):
        children = children_by_parent.get(b['id'], [])
        counts = Counter(c.get('status', 'unknown') for c in children)
        parts = ', '.join(f"{k}:{v}" for k, v in sorted(counts.items())) if counts else 'none'
        print(f"{b['id']}\t{b.get('status','')}\t{sum(counts.values())}\t{parts}\t{b.get('title','')}")
PY
```
