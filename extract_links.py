import re
import sys
import os

with open('/tmp/chunk_ac', 'r') as f:
    files = f.read().splitlines()

for file in files:
    if not file:
        continue
    md_path = os.path.join('.graph', file)
    if not os.path.exists(md_path):
        continue
    with open(md_path, 'r') as f:
        content = f.read()
    
    # Extract [[target|alias]] or [[target]]
    links = re.findall(r'\[\[(.*?)\]\]', content)
    for link in links:
        target = link.split('|')[0]
        # target might have #section
        target_file = target.split('#')[0]
        print(f"{file}\t{target_file}")
