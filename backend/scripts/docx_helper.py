#!/usr/bin/env python3
"""DOCX helper for translation: extract paragraphs and replace with translations.

Usage:
    python3 docx_helper.py extract <input.docx>
    python3 docx_helper.py replace <input.docx> <output.docx> <translations.json>

Extract outputs JSON to stdout: [{"index": 0, "text": "..."}, ...]
Replace reads translations JSON file: {"0": "translated", "1": "译文", ...}
Errors go to stderr with non-zero exit code.
"""

import sys
import json
import docx
from docx.oxml.ns import qn


def iter_paragraphs(body_element):
    """Yield (paragraph_element, sequential_index) in exact XML document order.

    Traverses direct children of w:body. For w:p elements, yields the paragraph.
    For w:tbl elements, descends into rows -> cells -> paragraphs.
    This ensures extract and replace produce identical indexing.
    """
    idx = 0
    for child in body_element:
        tag = child.tag.split('}')[-1] if '}' in child.tag else child.tag
        if tag == 'p':
            yield child, idx
            idx += 1
        elif tag == 'tbl':
            for row in child.iter(qn('w:tr')):
                for cell in row.iter(qn('w:tc')):
                    for p in cell.iter(qn('w:p')):
                        yield p, idx
                        idx += 1


def get_paragraph_text(p_element):
    """Extract full text from a paragraph element by joining all w:t texts."""
    texts = []
    for t in p_element.iter(qn('w:t')):
        if t.text:
            texts.append(t.text)
    return ''.join(texts)


def extract(input_path):
    """Extract all paragraph texts from a DOCX file."""
    doc = docx.Document(input_path)
    body = doc.element.body

    results = []
    for p_elem, idx in iter_paragraphs(body):
        text = get_paragraph_text(p_elem)
        results.append({"index": idx, "text": text})

    json.dump(results, sys.stdout, ensure_ascii=False)
    sys.stdout.write('\n')


def _set_run_text(run_element, text):
    """Set the text content of a run element, creating w:t if needed."""
    from docx.oxml import OxmlElement
    t_elems = run_element.findall(qn('w:t'))
    if t_elems:
        t_elems[0].text = text
        t_elems[0].set(qn('xml:space'), 'preserve')
    else:
        new_t = OxmlElement('w:t')
        new_t.text = text
        new_t.set(qn('xml:space'), 'preserve')
        run_element.append(new_t)


def replace(input_path, output_path, translations_path):
    """Replace paragraph texts in a DOCX file using translations from a JSON file."""
    with open(translations_path, 'r', encoding='utf-8') as f:
        translations = json.load(f)

    doc = docx.Document(input_path)
    body = doc.element.body

    for p_elem, idx in iter_paragraphs(body):
        key = str(idx)
        if key not in translations:
            continue

        translated_text = translations[key]
        runs = p_elem.findall(qn('w:r'))

        if not runs:
            # No runs exist — create one
            from docx.oxml import OxmlElement
            new_run = OxmlElement('w:r')
            new_t = OxmlElement('w:t')
            new_t.text = translated_text
            new_t.set(qn('xml:space'), 'preserve')
            new_run.append(new_t)
            p_elem.append(new_run)
        elif len(runs) == 1:
            # Single run — just replace text
            _set_run_text(runs[0], translated_text)
        else:
            # Multiple runs — write to first run, clear the rest
            _set_run_text(runs[0], translated_text)
            for run in runs[1:]:
                for t in run.findall(qn('w:t')):
                    t.text = ''

    doc.save(output_path)


def main():
    if len(sys.argv) < 2:
        print("Usage: docx_helper.py <extract|replace> ...", file=sys.stderr)
        sys.exit(1)

    command = sys.argv[1]

    try:
        if command == 'extract':
            if len(sys.argv) != 3:
                print("Usage: docx_helper.py extract <input.docx>", file=sys.stderr)
                sys.exit(1)
            extract(sys.argv[2])
        elif command == 'replace':
            if len(sys.argv) != 5:
                print("Usage: docx_helper.py replace <input.docx> <output.docx> <translations.json>", file=sys.stderr)
                sys.exit(1)
            replace(sys.argv[2], sys.argv[3], sys.argv[4])
        else:
            print(f"Unknown command: {command}", file=sys.stderr)
            sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
